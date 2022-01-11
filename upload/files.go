package upload

import (
	"database/sql"
	"fmt"
)

type File struct {
	Id           string `json:"Id"`
	Name         string `json:"name"`
	UploadedBy   string `json:"uploaded_by"`
	DateUploaded string `json:"date_uploaded"`
}

// CreateDB connects to our database, creates the files table that can store File data
func CreateDB(db *sql.DB) error {
	fmt.Println("creating table")
	_, err := db.Exec(`
			CREATE TABLE IF NOT EXISTS files (
			    id SERIAL NOT NULL PRIMARY KEY,
			    name VARCHAR(250) NOT NULL,
				uploaded_by VARCHAR(250) NOT NULL,
			    date_uploaded DATE NOT NULL
			);
		`)
	fmt.Println("created db")
	return err
}

// func (invoiceModel InvoiceModel) FindInvoicesBetween(from, to time.Time) ([]entities.Invoice, error) {
// 	rows, err := invoiceModel.Db.Query("select * from invoice where orderDate between ? and ?", from.Format("2006-01-02"), to.Format("2006-01-02"))
// 	if err != nil {
// 		return nil, err
// 	} else {
// 		invoices := []entities.Invoice{}
// 		for rows.Next() {
// 			var id int64
// 			var name string
// 			var orderDate string
// 			var status string
// 			err2 := rows.Scan(&id, &name, &orderDate, &status)
// 			if err2 != nil {
// 				return nil, err2
// 			} else {
// 				invoice := entities.Invoice{id, name, orderDate, status}
// 				invoices = append(invoices, invoice)
// 			}
// 		}
// 		return invoices, nil
// 	}
// }

// func getFile(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	key := vars["id"]
// 	for _, file := range File {
// 		if article.Id == key {
// 			json.NewEncoder(w).Encode(article)
// 		}
// 	}
// }

// func createArticle(w http.ResponseWriter, r *http.Request) {
// 	reqBody, _ := ioutil.ReadAll(r.Body)
// 	var article Article
// 	json.Unmarshal(reqBody, &article)
// 	Articles = append(Articles, article)
// 	json.NewEncoder(w).Encode(article)
// }
