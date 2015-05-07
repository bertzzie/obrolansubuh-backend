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
	err := dc.Trx.Select("name, email, password").From("contributors").
		Where("email = ?", email).LoadStruct(contributor)

	return contributor, err
}

func InitDB() {
	driver := revel.Config.StringDefault("db.driver", "mysql")
	conn := revel.Config.StringDefault("db.spec", "root:@/obrolansubuh")

	db, err := sql.Open(driver, conn)
	if err != nil {
		errMessage := fmt.Sprintf("[DBFatalError] Failed to open database.\nError Message: %s\n(driver: %s, spec: %s)",
			err.Error(), driver, conn)
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
		panic("[DTFE] Database Error. Please contact web administrator.")
	}

	return nil
}

func (db *DBRController) Commit() revel.Result {
	if db.Trx == nil {
		return nil
	}
	if err := db.Trx.Commit(); err != nil {
		errMessage := fmt.Sprintf("[DBTransactionError] Transaction commit. Error: %s", err.Error())
		revel.ERROR.Panicln(errMessage)
		panic("[DTFE] Database Error. Please contact web administrator.")
	}

	db.Trx = nil
	return nil
}

func (db *DBRController) RollBack() revel.Result {
	if db.Trx == nil {
		return nil
	}
	if err := db.Trx.Rollback(); err != nil && err != sql.ErrTxDone {
		errMessage := fmt.Sprintf("[DBTransactionError] Rollback failed. Error: %s", err.Error())
		revel.ERROR.Panicln(errMessage)
		panic("[DTFE] Database Error. Please contact web administrator.")
	}

	db.Trx = nil
	return nil
}
