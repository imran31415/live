package repo

import (
	"admin/models"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

// max amount of items we will fetch from the DB.
// TODO: replace this with some form of pagination.
const (
	globalLimit = 500
)

// SqlRepo contains logic to wrap around the GORM library.
type SqlRepo struct {
	*gorm.DB
}

func NewSqlRepo(host, name, pass, user string) (*SqlRepo, error) {
	// TODO paramaterize instance when we have more environments
	DB, err := setupDatabase(host, name, pass, user)
	if err != nil {
		return nil, err
	}
	return &SqlRepo{DB: DB}, nil
}

func (m *SqlRepo) Close() error {
	return m.DB.Close()
}

func setupDatabase(host, name, pass, user string) (*gorm.DB, error) {

	const mysqlCaCertPem = `
-----BEGIN CERTIFICATE-----
MIIDfzCCAmegAwIBAgIBADANBgkqhkiG9w0BAQsFADB3MS0wKwYDVQQuEyQ5OTBj
ZjQwMi03MjdiLTRjNjAtYTJhOC05N2FkMTgyNzJkZTkxIzAhBgNVBAMTGkdvb2ds
ZSBDbG91ZCBTUUwgU2VydmVyIENBMRQwEgYDVQQKEwtHb29nbGUsIEluYzELMAkG
A1UEBhMCVVMwHhcNMjAwNTI4MDIwNjAxWhcNMzAwNTI2MDIwNzAxWjB3MS0wKwYD
VQQuEyQ5OTBjZjQwMi03MjdiLTRjNjAtYTJhOC05N2FkMTgyNzJkZTkxIzAhBgNV
BAMTGkdvb2dsZSBDbG91ZCBTUUwgU2VydmVyIENBMRQwEgYDVQQKEwtHb29nbGUs
IEluYzELMAkGA1UEBhMCVVMwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIB
AQCObS7tYFRJbSzYwIHkPkOfm9YT9z5Dkwk2JR9Hgd3vAsC/UYnKOD28/NQQVO8d
d0F2q1LWODtxTnCvsrrZJ7iK8LMRItqZQfdyE8sDZBHZt64AEd/RYOxGQ1MCDKq/
F5teMqZ4Vltplu95JfEPbKV815/RKzu1JB7fIB8hrNBpk1gDvARGXux4gKQjOogs
a1mgusRIQ6v1aifEGPGY3uf30Sg/eXbBGjsNVNV7K0Vx+laweM2pcC7NGz1kPho7
HbCYCownmLOJL14FQG0zPQ/RTztt4x2O/blVyUAfR0ezODmIdXlo/eoyCxnsNxCA
qKFaLPxhNhM0mJL/D1ASiNxlAgMBAAGjFjAUMBIGA1UdEwEB/wQIMAYBAf8CAQAw
DQYJKoZIhvcNAQELBQADggEBAEIgUCwuvRWGUoN/aP2+W1UD3yocylLBdjoQKoNF
hAoi2yiCX4MvAc13uyjINo041h9cIygprnZ9Jmd+t2ib7TkT2HUM8QuZjkwBRqHA
oTwbYuIGitdOgZS7YU73xRlnTUQbkTdHNmS+hJ/9FL533COTy5+ibRgVEYDBKDAY
o/fgamo3gGe1zbQn2XhwU2zh/9ebc9XuIRgH/lh1yjy95KFcX2aDY1C3d5dtFpmD
bXpUFKKkg1L3FO61Opu0rb0NdI+uCm6mPZjGls/JkKAndcvpNkO2gQiGV3o5T9c8
54NuyO4pUMCSmyIq3OtDaxEKi4h2WiljjeWY7qu2cv7GdPw=
-----END CERTIFICATE-----`

	const mysqlClientKeyPem = `
-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEAmLP/CEYkvmYyK/AKM+G5DhL0CGLOyiYU4IDMtUx/SHTkDLt6
XOFftNDGAFqFHcu3X1bOT/elCzJULoi3XIdwTMqkbquqQcBmTmoIhQDFV/ECJZ9S
RWyfkQWd5/JZUom3lSX9l+XrV9QTYDCWH6lfDE3KX5WbPiajZbF/xTfy5rKANuZt
x2vg1Eg6vUsaRR+ave/yNTe96PJ7uiteHFkgcCUdOueP3jjMJksc0NQ7R1sfHjIT
iGEJanLVEBoAf4O3IqkG0IZAtgxyb/0s8H40gIaQFIgSoPxAYFtJxBji/XUSrMNe
6BdCp+7j+kdg1dAth58LB2+ELbsemVwnz6Q9SwIDAQABAoIBAQCFswt4Dho7zrTc
/YWhWWAyr2uVMBYIroEW9A7IKltDcHz/nOKNEhM++JB0XM0yglxikFmBcL3D6OQF
/lQ1Iqyzv7Vq5MjkWvX4cCRXd45R6kXL9QwOlwW67yULoYiNmODxKNs5tOhy9M+m
J8Q5oo0C89VeDVpod2IXNus9HTiCvaPZwp8CGnzANrY+Ybi3wO8wKnr6yABZu37/
1lZn2bOf/459eG+b59WbS8jvxjNbl7FIC6m6HTONzMtYUBJlOOogysVjoiO4L2vk
W5itCS/hkIXlroLeXDZZY0X4tOM+y78KC4On9FO6zPI3Zg7o00bgCD5+1sKgFhZd
z12zXd2JAoGBAPI/elmoVbYMTTL/K5Mg7QLx7tIiqsqtda8kG3r8clY9P54+H/TY
olg5Hiz3tVu0WIIQ87yB8y/LLrSGjWsZg1MR5yhYDlrwaWJBnSyLzxsvXWSSg0rt
dCsOOlR1FHRwv1T7YxCA8TW8iPFjWUpdxAA3mWp/Iw95VF2nr0O1EJo9AoGBAKFf
MB0l6QuLMkGGwdsSXMstsDnXI1thciIVResQqS9Ci9iqgdnhq+Dx4EjQpLAtXH01
Fq3WDxCD6WPuS3dYWAHDnZVmXJHh+sq+qwmstslI8zUJWofD9CKphmHb95i5iKHn
WQtfZjT4EMwQVAHacWzuTaHAAK1DPA96iOBTgZYnAoGBAJoH4Lz3eyBZLBEcDNHt
Yqa3vHnizyQ2LRki5VJLCExrf3MX32vo/zkHgHdpPejEgG6bZs9a9Y1TLSxeTbdm
roj4XjnZ267ZJLj0LYMwloybjk+vlUnkODRURKSFGW98bTwU6AWLZ1QawBx0ZkcR
3dmhgKwlkN568DjosVlk3NylAoGAK9BufduHNO0sTgJKrDKGI1xaVroFDZCdrodc
HoC9julgkwlojEHrqv3BScPskzEdxZkeeUB/gppuSgWvU84WxxPXu3K5e5qBv36Z
bd0JHAnEjwflHqujo62noPZaeYsWf+8SjDXwyDz6Qo3EYWRwG4VwapR5GpIAwqsg
ctf5fU0CgYEA5Bw7yUvqYXDmJ4xpPmtvyNy+k/ZBSMfGiENLej1rG6w6kj4bddq+
m3glSpIoCT0sVCbpXVlnQM4GFowp5Tw2FzdatKXT5Gch3YWNxxjnaGRyWm3SMQNq
k/IL8l2aA7OgTA2R3Vszf6e4Rm80CZMEvAtrKQFp/jLbcXd/OGMJTQA=
-----END RSA PRIVATE KEY-----`

	const mysqlClientCertPem = `
-----BEGIN CERTIFICATE-----
MIIDZTCCAk2gAwIBAgIEIFYXKjANBgkqhkiG9w0BAQsFADB/MS0wKwYDVQQuEyQ0
ZWVlYzJiNC03ZTczLTRkYzEtYmJmZS05YmQ5NmUzMzA5NWUxKzApBgNVBAMTIkdv
b2dsZSBDbG91ZCBTUUwgQ2xpZW50IENBIGJhY2tlbmQxFDASBgNVBAoTC0dvb2ds
ZSwgSW5jMQswCQYDVQQGEwJVUzAeFw0yMDA4MDEyMTI2NTdaFw0zMDA3MzAyMTI3
NTdaMDUxEDAOBgNVBAMTB2JhY2tlbmQxFDASBgNVBAoTC0dvb2dsZSwgSW5jMQsw
CQYDVQQGEwJVUzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAJiz/whG
JL5mMivwCjPhuQ4S9AhizsomFOCAzLVMf0h05Ay7elzhX7TQxgBahR3Lt19Wzk/3
pQsyVC6It1yHcEzKpG6rqkHAZk5qCIUAxVfxAiWfUkVsn5EFnefyWVKJt5Ul/Zfl
61fUE2Awlh+pXwxNyl+Vmz4mo2Wxf8U38uaygDbmbcdr4NRIOr1LGkUfmr3v8jU3
vejye7orXhxZIHAlHTrnj944zCZLHNDUO0dbHx4yE4hhCWpy1RAaAH+DtyKpBtCG
QLYMcm/9LPB+NICGkBSIEqD8QGBbScQY4v11EqzDXugXQqfu4/pHYNXQLYefCwdv
hC27HplcJ8+kPUsCAwEAAaMzMDEwCQYDVR0TBAIwADAkBgNVHREEHTAbgRluZWxz
b25hcG9sbG9sZWVAZ21haWwuY29tMA0GCSqGSIb3DQEBCwUAA4IBAQAM5mfHYB5j
ZclgWBNh0lEpRLunQ9w/qXh17NReiaMjY31GXkQ6gNLT0XcqxwHYkvhDFFcMRMQc
ODjIzrsjvXPuA/87HaPt+BGpOklECkOrT82Zn3DTrKDKk2NKiGBKaUJ2KOGv0WY/
76BpZYXnGJ/s8NQtmagPGKHAI+73MkKOO6kA9saxykV/Kpv/HNkJL2yKe5MjJmkT
f9OS2VqtZSJigYC5WxOeS9anRHFebV3RhSC8yX2fLEbHJQ5V4+LslNJRTZlZREBU
zlU6IbOQX7aAYSSgfeIuApOLKvqOxaLHfoAI/jgvAhoTu4lz2mwA1v9UHV2taoIw
FwUHIz5I+5JQ
-----END CERTIFICATE-----`

	rootCertPool := x509.NewCertPool()
	if ok := rootCertPool.AppendCertsFromPEM([]byte(mysqlCaCertPem)); !ok {
		log.Fatal("Failed to append CaCertPem.")
	}

	clientCert := make([]tls.Certificate, 0, 1)
	certs, err := tls.X509KeyPair([]byte(mysqlClientCertPem), []byte(mysqlClientKeyPem))
	if err != nil {
		log.Fatal("Failed to create KeyPair")
		log.Fatal(err)
	}
	clientCert = append(clientCert, certs)
	mysql.RegisterTLSConfig("custom", &tls.Config{
		RootCAs:      rootCertPool,
		Certificates: clientCert,
		ServerName:   "livehub-277906:prod",
	})

	// try to connect to mysql database.
	cfg := mysql.Config{
		User:                 user,
		Passwd:               pass,
		Addr:                 host, //IP:PORT
		Net:                  "tcp",
		DBName:               name,
		Loc:                  time.Local,
		AllowNativePasswords: true,
		ParseTime:            true,
	}

	cfg.TLSConfig = "custom"

	str := cfg.FormatDSN() + "&charset=utf8"

	log.Println("Attempting to connect to DB", str)
	DB, err := gorm.Open("mysql", str)

	if err != nil {
		log.Println("Failed to connect, err = ", err)
		log.Println("Trying local DB")
		connectionString := fmt.Sprintf("%s:%s@(%s)/%s?charset=utf8&parseTime=True&loc=Local", user, pass, host, name)
		if DB, err = gorm.Open("mysql", connectionString); err != nil {
			log.Println("Unable to connect to local DB, please check DB connection")
			return nil, err
		} else {
			log.Println("Successfully connected to local mysql DB.  Note this is the local mysql instance to wherever this server is running. ")
		}
	} else {
		log.Println("Successfully connected to DB for connection: ", str)
	}

	// setup tables
	if err = DB.AutoMigrate(
		&models.User{},
		&models.Customer{},
		&models.Order{},
		&models.Image{},
		&models.ZoomToken{},
		&models.Session{},
	).Error; err != nil {

		return nil, err
	}
	return DB, nil

}
