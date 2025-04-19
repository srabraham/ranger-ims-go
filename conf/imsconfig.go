package conf

import "github.com/google/uuid"

var Cfg *IMSConfig

func DefaultIMS() *IMSConfig {
	return &IMSConfig{
		Core: ConfigCore{
			Host:       "localhost",
			Port:       80,
			JWTSecret:  uuid.New().String(),
			Deployment: "dev",
		},
	}
}

type IMSConfig struct {
	Core             ConfigCore
	AttachmentsStore struct {
		S3 struct {
			S3AccessKeyId     string
			S3SecretAccessKey string
			S3DefaultRegion   string
			S3Bucket          string
		}
	}
	Store struct {
		SQLite struct {
			File string
		}
		MySQL struct {
			HostName string
			HostPort int32
			Database string
			Username string
			Password string
		}
	}
	Directory struct {
		File struct {
			File string
		}
		ClubhouseDB struct {
			Hostname string
			HostPort int32
			Database string
			Username string
			Password string
		}
	}
}

type ConfigCore struct {
	ServerRoot       string
	TokenLifetime    int64
	Dataroot         string
	Admins           []string
	DataStore        string
	Directory        string
	ConfigRoot       string
	CachedResources  string
	Host             string
	Port             int32
	MasterKey        string
	JWTSecret        string
	AttachmentsStore string
	Deployment       string
}
