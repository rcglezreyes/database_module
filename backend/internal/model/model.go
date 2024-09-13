package model

import (
	"archive/zip"
	"backend/internal/client"
	"backend/internal/config"
	"backend/internal/entity"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

type model struct {
	client        client.MongoDBClient
	dbCredentials *entity.DBCredentials
	loggers       *entity.Loggers
}
type Model interface {
	LoadBatchData() error
	DownloadData() error
	GetFiles() ([]*entity.FileInfo, error)
	GetData(collection string) ([]interface{}, error)
	GetAllCountData(collections []string) (map[string]int64, error)
	ProcessDataPredictionAssessments() ([]entity.ProcessedPredictionAssessmentResult, error)
	ProcessDataVlePredictions() ([]entity.ProcessedPredictionVleResult, error)
	GetScoreDistributionPredictionAssessments() ([]entity.ScoreRangePredictionAssessments, error)
	GetAveragePredictedScoreByAssessmentType() ([]entity.AssessmentTypeAverage, error)
	GetStudentCountByAssessmentID() ([]entity.AssessmentStudentCount, error)
}

func NewModel(client client.MongoDBClient, loggers *entity.Loggers) Model {
	_, dbCredentials, _ := config.DBCredentials()
	return &model{
		client:        client,
		dbCredentials: &dbCredentials,
		loggers:       loggers,
	}
}

func (m *model) LoadBatchData() error {
	enviroment := viper.GetString(config.Envirornment)
	var filePathRead string
	if enviroment == "DEV" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			m.loggers.ErrorLogger.Printf("No se pudo obtener el directorio de inicio del usuario: %v", err)
		}
		filePathRead = filepath.Join(homeDir, viper.GetString(config.FilePathReadDev))
	} else {
		filePathRead = viper.GetString(config.FilePathReadQa)
	}
	files := []struct {
		Path         string
		Collection   string
		ProcessBatch func(string, [][]string) error
	}{
		{filePathRead + "/courses.csv", "courses", m.processBatch},
		{filePathRead + "/assessments.csv", "assessments", m.processBatch},
		{filePathRead + "/studentInfo.csv", "studentInfo", m.processBatch},
		{filePathRead + "/vle.csv", "vle", m.processBatch},
		{filePathRead + "/studentAssessment.csv", "studentAssessment", m.processBatch},
		{filePathRead + "/studentVle.csv", "studentVle", m.processBatch},
		{filePathRead + "/studentRegistration.csv", "studentRegistration", m.processBatch},
	}
	for _, file := range files {
		m.loggers.InfoLogger.Printf("Procesando archivo: %s", file.Path)
		batchSize := viper.GetInt(config.BatchSize)
		if err := m.processCSVInBatches(file.Path, batchSize, func(batch [][]string) error {
			return file.ProcessBatch(file.Collection, batch)
		}); err != nil {
			m.loggers.ErrorLogger.Printf("Error al procesar archivo %s: %v", file.Path, err)
		}
	}

	m.loggers.InfoLogger.Println("Procesamiento completado.")
	return nil

}

func (m *model) DownloadData() error {
	files, err := m.GetFiles()
	if err != nil {
		m.loggers.ErrorLogger.Printf("Error al obtener archivos: %v", err)
		return err
	}
	if len(files) == 0 {
		url := viper.GetString(config.UrlOulad)
		m.loggers.InfoLogger.Printf("Descargando archivo de: %s", url)
		var pathDownload, zipPath, extractPath string

		environment := viper.GetString(config.Envirornment)
		if environment == "DEV" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				m.loggers.ErrorLogger.Printf("No se pudo obtener el directorio de inicio del usuario: %v", err)
			}
			pathDownload = filepath.Join(homeDir, viper.GetString(config.FilePathDownloadDev))
			zipPath = filepath.Join(pathDownload, viper.GetString(config.FileNameZip))
			extractPath = filepath.Join(homeDir, viper.GetString(config.FilePathReadDev))
		} else {
			zipPath = viper.GetString(config.FilePathDownloadQa) + "/" + viper.GetString(config.FileNameZip)
			extractPath = viper.GetString(config.FilePathDownloadQa)
		}

		log.Println("Descargando archivo...")
		err := m.downloadZip(url, zipPath)
		if err != nil {
			m.loggers.ErrorLogger.Printf("Error al descargar archivo: %v", err)
			return err
		}
		m.loggers.InfoLogger.Println("Archivo descargado exitosamente.")

		log.Println("Descomprimiendo archivo...")
		err = m.unzipFile(zipPath, extractPath)
		if err != nil {
			m.loggers.ErrorLogger.Printf("Error al descomprimir archivo: %v", err)
			return err
		}
		log.Println("Archivo descomprimido exitosamente.")
		return nil
	}
	m.loggers.InfoLogger.Println("No hay archivos para descargar.")
	return nil
}

func (m *model) processCSVInBatches(filePath string, batchSize int, processBatch func([][]string) error) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	var batch [][]string
	for {
		record, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				if len(batch) > 0 {
					if err := processBatch(batch); err != nil {
						return err
					}
				}
				break
			} else {
				return err
			}
		}
		batch = append(batch, record)
		if len(batch) >= batchSize {
			if err := processBatch(batch); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}
	return nil
}
func (m *model) processBatch(collectionName string, batch [][]string) error {
	var data []interface{}
	m.loggers.InfoLogger.Printf("Procesando lote de %d registros para la colección %s", len(batch)-1, collectionName)

	for _, record := range batch[1:] {
		switch collectionName {
		case "courses":
			length, _ := strconv.Atoi(record[2])
			value := entity.Courses{
				CodeModule:       record[0],
				CodePresentation: record[1],
				Length:           length,
			}
			data = append(data, value)
		case "assessments":
			idAssessment, _ := strconv.Atoi(record[2])
			date, _ := strconv.Atoi(record[4])
			weight, _ := strconv.Atoi(record[5])
			value := entity.Assessments{
				IdAssessment:     idAssessment,
				CodeModule:       record[0],
				CodePresentation: record[1],
				AssessmentType:   record[3],
				Date:             date,
				Weight:           weight,
			}
			data = append(data, value)
		case "vle":
			idSite, _ := strconv.Atoi(record[0])
			weekFrom, _ := strconv.Atoi(record[4])
			weekTo, _ := strconv.Atoi(record[5])
			value := entity.Vle{
				IdSite:           idSite,
				CodeModule:       record[1],
				CodePresentation: record[2],
				ActivityType:     record[3],
				WeekFrom:         weekFrom,
				WeekTo:           weekTo,
			}
			data = append(data, value)
		case "studentInfo":
			idStudent, _ := strconv.Atoi(record[2])
			numOfPrevAttempts, _ := strconv.Atoi(record[8])
			studiedCredits, _ := strconv.Atoi(record[9])
			value := entity.StudentInfo{
				IdStudent:         idStudent,
				CodeModule:        record[0],
				CodePresentation:  record[1],
				Gender:            record[3],
				Region:            record[4],
				HighestEducation:  record[5],
				IMDBand:           record[6],
				AgeBand:           record[7],
				NumOfPrevAttempts: numOfPrevAttempts,
				StudiedCredits:    studiedCredits,
				Disability:        record[10],
				FinalResult:       record[11],
			}
			data = append(data, value)
		case "studentRegistration":
			idStudent, _ := strconv.Atoi(record[2])
			dateRegistration, _ := strconv.Atoi(record[3])
			dateUnregistration, _ := strconv.Atoi(record[4])
			value := entity.StudentRegistration{
				CodeModule:         record[0],
				CodePresentation:   record[1],
				IdStudent:          idStudent,
				DateRegistration:   dateRegistration,
				DateUnregistration: dateUnregistration,
			}
			data = append(data, value)
		case "studentAssessment":
			idAssessment, _ := strconv.Atoi(record[0])
			idStudent, _ := strconv.Atoi(record[1])
			dateSubmitted, _ := strconv.Atoi(record[2])
			isBanked, _ := strconv.Atoi(record[3])
			score, _ := strconv.ParseFloat(record[4], 64)
			value := entity.StudentAssessment{
				IdAssessment:  idAssessment,
				IdStudent:     idStudent,
				DateSubmitted: dateSubmitted,
				IsBanked:      isBanked,
				Score:         score,
			}
			data = append(data, value)
		case "studentVle":
			idStudent, _ := strconv.Atoi(record[2])
			idSite, _ := strconv.Atoi(record[3])
			date, _ := strconv.Atoi(record[4])
			sumClick, _ := strconv.Atoi(record[5])
			value := entity.StudentVle{
				CodeModule:       record[0],
				CodePresentation: record[1],
				IdStudent:        idStudent,
				IdSite:           idSite,
				Date:             date,
				SumClick:         sumClick,
			}
			data = append(data, value)
		}
	}
	batchSize := viper.GetInt(config.BatchSize)
	err := m.client.BatchInsert(m.dbCredentials.Dbname, collectionName, data, batchSize)
	return err
}

func (m *model) downloadZip(url, filepath string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
func (m *model) unzipFile(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

func (m *model) GetFiles() ([]*entity.FileInfo, error) {
	var dirPath string
	environment := viper.GetString(config.Envirornment)
	if environment == "DEV" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			m.loggers.ErrorLogger.Printf("No se pudo obtener el directorio de inicio del usuario: %v", err)
		}
		dirPath = filepath.Join(homeDir, viper.GetString(config.FilePathReadDev))
	} else {
		dirPath = viper.GetString(config.FilePathDownloadQa)
	}
	m.loggers.InfoLogger.Printf("Obteniendo archivos de: %s", dirPath)

	var files []*entity.FileInfo

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".csv") {
			files = append(files, &entity.FileInfo{FileName: info.Name(), FileSize: m.formatFileSize(info.Size())})
		}
		return nil
	})
	m.loggers.InfoLogger.Printf("Se encontraron %d archivos", len(files))
	if err != nil {
		m.loggers.ErrorLogger.Printf("Error al obtener archivos: %v", err)
		return nil, err
	}
	return files, nil
}
func (m *model) GetAllCountData(collections []string) (map[string]int64, error) {
	return m.client.GetAllCountData(m.dbCredentials.Dbname, collections)
}

func (m *model) GetData(collection string) ([]interface{}, error) {
	return m.client.GetData(m.dbCredentials.Dbname, collection)
}

func (m *model) ProcessDataPredictionAssessments() ([]entity.ProcessedPredictionAssessmentResult, error) {
	return m.client.ProcessDataPredictionAssessments(m.dbCredentials.Dbname)
}

func (m *model) ProcessDataVlePredictions() ([]entity.ProcessedPredictionVleResult, error) {
	return m.client.ProcessDataVlePredictions(m.dbCredentials.Dbname)
}

func (m *model) GetScoreDistributionPredictionAssessments() ([]entity.ScoreRangePredictionAssessments, error) {
	return m.client.GetScoreDistributionPredictionAssessments(m.dbCredentials.Dbname)
}

func (m *model) GetAveragePredictedScoreByAssessmentType() ([]entity.AssessmentTypeAverage, error) {
	return m.client.GetAveragePredictedScoreByAssessmentType(m.dbCredentials.Dbname)
}

func (m *model) GetStudentCountByAssessmentID() ([]entity.AssessmentStudentCount, error) {
	return m.client.GetStudentCountByAssessmentID(m.dbCredentials.Dbname)
}

func (m *model) formatFileSize(size int64) string {
	const (
		KB = 1 << (10 * (iota + 1))
		MB
		GB
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d B", size)
	}
}
