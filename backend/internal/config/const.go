package config

const (
	//General
	CONFIG_FILE_NAME string = "config"
	CONFIG_FILE_TYPE string = "json"
	APP_PORT         string = "APP_PORT"
	STATUS_OK        string = "Success"
	STATUS_FAILED    string = "Failed"
	//Database
	HostDev     string = "DB_HOST_DEV"
	PortDev     string = "DB_PORT_DEV"
	UserDev     string = "DB_USER_DEV"
	PasswordDev string = "DB_PASSWORD_DEV"
	DbnameDev   string = "DB_DATABASE_DEV"
	HostQa      string = "DB_HOST_QA"
	PortQa      string = "DB_PORT_QA"
	UserQa      string = "DB_USER_QA"
	PasswordQa  string = "DB_PASSWORD_QA"
	DbnameQa    string = "DB_DATABASE_QA"
	//URLsExternas
	UrlOulad            string = "URL_OULAD"
	FilePathDownloadDev string = "FILE_PATH_DOWNLOAD_DEV"
	FilePathDownloadQa  string = "FILE_PATH_DOWNLOAD_QA"
	FileNameZip         string = "FILE_NAME_ZIP"
	FilePathReadDev     string = "FILE_PATH_READ_DEV"
	FilePathReadQa      string = "FILE_PATH_READ_QA"
	Envirornment        string = "ENVIRONMENT"
	//MongoDB
	BatchSize string = "BATCH_SIZE"
)
