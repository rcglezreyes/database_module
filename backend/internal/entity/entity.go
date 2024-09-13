package entity

import (
	"log"
	"time"
)

// Courses estructura para el archivo courses.csv
type Courses struct {
	CodeModule       string `json:"code_module"`
	CodePresentation string `json:"code_presentation"`
	Length           int    `json:"length"`
}

// Assessments estructura para el archivo assessments.csv
type Assessments struct {
	IdAssessment     int    `json:"id_assessment"`
	CodeModule       string `json:"code_module"`
	CodePresentation string `json:"code_presentation"`
	AssessmentType   string `json:"assessment_type"`
	Date             int    `json:"date"`
	Weight           int    `json:"weight"`
}

// Vle estructura para el archivo vle.csv
type Vle struct {
	IdSite           int    `json:"id_site"`
	CodeModule       string `json:"code_module"`
	CodePresentation string `json:"code_presentation"`
	ActivityType     string `json:"activity_type"`
	WeekFrom         int    `json:"week_from"`
	WeekTo           int    `json:"week_to"`
}

// StudentInfo estructura para el archivo studentInfo.csv
type StudentInfo struct {
	IdStudent         int    `json:"id_student"`
	CodeModule        string `json:"code_module"`
	CodePresentation  string `json:"code_presentation"`
	Gender            string `json:"gender"`
	Region            string `json:"region"`
	HighestEducation  string `json:"highest_education"`
	IMDBand           string `json:"imd_band"`
	AgeBand           string `json:"age_band"`
	NumOfPrevAttempts int    `json:"num_of_prev_attempts"`
	StudiedCredits    int    `json:"studied_credits"`
	Disability        string `json:"disability"`
	FinalResult       string `json:"final_result"`
}

// StudentRegistration estructura para el archivo studentRegistration.csv
type StudentRegistration struct {
	CodeModule         string `json:"code_module"`
	CodePresentation   string `json:"code_presentation"`
	IdStudent          int    `json:"id_student"`
	DateRegistration   int    `json:"date_registration"`
	DateUnregistration int    `json:"date_unregistration"`
}

// StudentAssessment estructura para el archivo studentAssessment.csv
type StudentAssessment struct {
	IdAssessment  int     `json:"id_assessment"`
	IdStudent     int     `json:"id_student"`
	DateSubmitted int     `json:"date_submitted"`
	IsBanked      int     `json:"is_banked"`
	Score         float64 `json:"score"`
}

// StudentVle estructura para el archivo studentVle.csv
type StudentVle struct {
	CodeModule       string `json:"code_module"`
	CodePresentation string `json:"code_presentation"`
	IdStudent        int    `json:"id_student"`
	IdSite           int    `json:"id_site"`
	Date             int    `json:"date"`
	SumClick         int    `json:"sum_click"`
}
type DBCredentials struct {
	Host     string
	Port     int
	User     string
	Password string
	Dbname   string
}

type MongoDBCredentials struct {
	URI string
}
type Loggers struct {
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
}
type ResponseGeneric struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
type FileInfo struct {
	FileName string `json:"file_name"`
	FileSize string `json:"file_size"`
}
type CollectionsRequest struct {
	Collections []string `json:"collections"`
}
type PredictionAssessment struct {
	StudentID      int       `bson:"student_id"`
	AssessmentID   int       `bson:"assessment_id"`
	PredictedScore float64   `bson:"predicted_score"`
	PredictionDate time.Time `bson:"prediction_date"`
}
type ProcessedPredictionAssessmentResult struct {
	StudentID      int     `bson:"student_id"`
	AssessmentID   int     `bson:"assessment_id"`
	PredictedScore float64 `bson:"predicted_score"`
}

type PredictionVle struct {
	StudentID    int    `bson:"id_student"`
	ResourceType string `bson:"resource_type"`
	Clicks       int    `bson:"sum_click"`
}

// Estructura para almacenar el resultado del procesamiento
type ProcessedPredictionVleResult struct {
	StudentID      int     `bson:"student_id"`
	PredictedScore float64 `bson:"predicted_score"`
}

type ScoreRangePredictionAssessments struct {
	Range        string `json:"range"`
	StudentCount int    `json:"student_count"`
}

type AssessmentTypeAverage struct {
    AssessmentType string  `bson:"assessment_type"`
    AverageScore   float64 `bson:"average_score"`
}

type AssessmentStudentCount struct {
	AssessmentID int `bson:"_id"`
	StudentCount int    `bson:"student_count"`
}
