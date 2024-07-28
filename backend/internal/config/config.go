package config

import (
	"backend/internal/entity"
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

func DBCredentials() (entity.MongoDBCredentials, error) {
	port, err := strconv.Atoi(viper.GetString(Port))
	if err != nil {
		return entity.MongoDBCredentials{}, err
	}
	dbCredentials := entity.DBCredentials{
		Host:     viper.GetString(Host),
		Port:     port,
		User:     viper.GetString(User),
		Password: viper.GetString(Password),
		Dbname:   viper.GetString(Dbname),
	}
	return entity.MongoDBCredentials{
		URI: "mongodb://" + dbCredentials.User + ":" + dbCredentials.Password + "@" + dbCredentials.Host + ":" + strconv.Itoa(dbCredentials.Port),
	}, nil
}
