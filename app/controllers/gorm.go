package controllers

import (
	"database/sql"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/revel/revel"
	"obrolansubuh.com/models"
)

var ORM *gorm.DB

type GormController struct {
	*revel.Controller
	Trx *gorm.DB
}

func InitDB() {
	driver := revel.Config.StringDefault("db.driver", "mysql")
	conn := revel.Config.StringDefault("db.spec", "root:@/obrolansubuh")

	dbm, err := gorm.Open(driver, conn)

	if err != nil {
		errMessage := fmt.Sprintf("[DBFatalError] Failed to open database (driver: %s, spec: %s).\nError Message: %s\n",
			driver, conn, err.Error())
		revel.ERROR.Panicln(errMessage)
		panic("[DBFE] Database Connection Error. Please contact web administrator.")
	}

	ORM = &dbm

	dbm.AutoMigrate(&models.Contributor{}, &models.Post{}, &models.ContributorType{}, &models.SiteInfo{})

	siteInfo := models.SiteInfo{}

	dbm.FirstOrCreate(&siteInfo, models.SiteInfo{
		AboutUsTitle:   "About Us",
		AboutUsContent: "This is ObrolanSubuh.com",
		TwitterURL:     "obrolansubuh",
		FacebookURL:    "obrolansubuh",
	})

	// If there's no user, create default admin user
	count := 0
	if dbm.Model(&models.Contributor{}).Count(&count); count < 1 {
		typeAdmin := models.ContributorType{Type: "ADMIN"}
		typeWriter := models.ContributorType{Type: "WRITER"}

		dbm.Create(&typeAdmin)
		dbm.Create(&typeWriter)

		admin := models.Contributor{
			Name:  "Default Admin",
			Email: "admin@obrolansubuh.com",
			About: "Default Admin ObrolanSubuh.com",
			Photo: "/public/img/default-user.png",
			Type:  &typeAdmin,
		}
		admin.SetPassword("admin@obrolansubuh.com")

		writer := models.Contributor{
			Name:  "Default Writer",
			Email: "writer@obrolansubuh.com",
			About: "Default Writer ObrolanSubuh.com",
			Photo: "/public/img/default-user.png",
			Type:  &typeWriter,
		}
		writer.SetPassword("writer@obrolansubuh.com")

		dbm.Create(&admin)
		dbm.Create(&writer)
	}
}

func (gc *GormController) GetContributor(email string) (*models.Contributor, error) {
	contributor := &models.Contributor{}

	tx := gc.Trx.Preload("Type").Where("email = ?", email).First(&contributor)

	return contributor, tx.Error
}

func (gc *GormController) Begin() revel.Result {
	trx := ORM.Begin()
	if err := trx.Error; err != nil {
		errMessage := fmt.Sprintf("[DBTransactionError] Begin transaction error: %s", err.Error())
		revel.ERROR.Panicln(errMessage)
		panic(gc.Message("errors.db.generic"))
	}

	gc.Trx = trx

	return nil
}

func (gc *GormController) Commit() revel.Result {
	if gc.Trx == nil {
		return nil
	}

	gc.Trx.Commit()
	if err := gc.Trx.Error; err != nil && err != sql.ErrTxDone {
		errMessage := fmt.Sprintf("[DBTransactionError] Transaction commit error: %s", err.Error())
		revel.ERROR.Panicln(errMessage)
		panic(gc.Message("errors.db.generic"))
	}

	gc.Trx = nil
	return nil
}

func (gc *GormController) RollBack() revel.Result {
	if gc.Trx == nil {
		return nil
	}

	gc.Trx.Rollback()
	if err := gc.Trx.Error; err != nil && err != sql.ErrTxDone {
		errMessage := fmt.Sprintf("[DBTransactionError] Rollback failed error: %s", err.Error())
		revel.ERROR.Panicln(errMessage)
		panic(gc.Message("errors.db.generic"))
	}

	gc.Trx = nil
	return nil
}
