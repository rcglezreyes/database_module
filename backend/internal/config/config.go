package config

import (
	"backend/internal/entity"
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/spf13/viper"
)

func ConfigEnv() (error, bool) {
	viper.SetConfigName(CONFIG_FILE_NAME)
	viper.AddConfigPath(RootDir())
	viper.AutomaticEnv()
	viper.SetConfigType(CONFIG_FILE_TYPE)
	if err := viper.ReadInConfig(); err != nil {
		return err, false
	}
	return nil, true
}

func RootDir() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(b), "../..")
}

func DBCredentials() (entity.MongoDBCredentials, entity.DBCredentials, error) {
	var host, user, password, dbname string
	var port int
	if viper.GetString(Envirornment) == "DEV" {
		host = viper.GetString(HostDev)
		user = viper.GetString(UserDev)
		password = viper.GetString(PasswordDev)
		dbname = viper.GetString(DbnameDev)
		port, _ = strconv.Atoi(viper.GetString(PortDev))
	} else {
		host = viper.GetString(HostQa)
		user = viper.GetString(UserQa)
		password = viper.GetString(PasswordQa)
		dbname = viper.GetString(DbnameQa)
		port, _ = strconv.Atoi(viper.GetString(PortQa))
	}
	dbCredentials := entity.DBCredentials{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Dbname:   dbname,
	}
	uri := "mongodb://" +
		dbCredentials.User + ":" +
		dbCredentials.Password + "@" +
		dbCredentials.Host + ":" +
		strconv.Itoa(dbCredentials.Port) + "/" +
		dbCredentials.Dbname + "?authMechanism=SCRAM-SHA-1&authSource=admin"
	fmt.Println(uri)
	return entity.MongoDBCredentials{
		URI: uri,
	}, dbCredentials, nil
}
