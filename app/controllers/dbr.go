package controllers

import (
	"database/sql"
	"fmt"
	"github.com/gocraft/dbr"
	"github.com/revel/revel"
	"obrolansubuh.com/models"
)

var connection *dbr.Connection

type DBRController struct {
	*revel.Controller
	Trx *dbr.Tx
}

func (dc *DBRController) GetContributor(email string) (*models.Contributor, error) {
	contributor := &models.Contributor{}
	err := dc.Trx.Select("id, name, email, password, photo").From("contributors").
		Where("email = ?", email).LoadStruct(contributor)

	return contributor, err
}

func InitDB() {
	driver := revel.Config.StringDefault("db.driver", "mysql")
	conn := revel.Config.StringDefault("db.spec", "root:@/obrolansubuh")

	db, err := sql.Open(driver, conn)
	if err != nil {
		errMessage := fmt.Sprintf("[DBFatalError] Failed to open database (driver: %s, spec: %s).\nError Message: %s\n",
			driver, conn, err.Error())
		revel.ERROR.Panicln(errMessage)
		panic("[DBFE] Database Connection Error. Please contact web administrator.")
	}

	connection = dbr.NewConnection(db, nil)
}

func (dc *DBRController) Begin() revel.Result {
	var err error
	dc.Trx, err = connection.NewSession(nil).Begin()
	if err != nil {
		errMessage := fmt.Sprintf("[DBTransactionError] Begin transaction error: %s", err.Error())
		revel.ERROR.Panicln(errMessage)
		panic(dc.Message("errors.db.generic"))
	}

	return nil
}

func (dc *DBRController) Commit() revel.Result {
	if dc.Trx == nil {
		return nil
	}
	if err := dc.Trx.Commit(); err != nil {
		errMessage := fmt.Sprintf("[DBTransactionError] Transaction commit error: %s", err.Error())
		revel.ERROR.Panicln(errMessage)
		panic(dc.Message("errors.db.generic"))
	}

	dc.Trx = nil
	return nil
}

func (dc *DBRController) RollBack() revel.Result {
	if dc.Trx == nil {
		return nil
	}
	if err := dc.Trx.Rollback(); err != nil && err != sql.ErrTxDone {
		errMessage := fmt.Sprintf("[DBTransactionError] Rollback failed error: %s", err.Error())
		revel.ERROR.Panicln(errMessage)
		panic(dc.Message("errors.db.generic"))
	}

	dc.Trx = nil
	return nil
}
