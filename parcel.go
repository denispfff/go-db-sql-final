package main

import (
	"database/sql"
	"fmt"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	// реализуйте добавление строки в таблицу parcel, используйте данные из переменной p
	res, err := s.db.Exec("INSERT INTO parcel (client, status, address, created_At) VALUES (:client, :status, :address, :created_At)",
		sql.Named("client", p.Client),
		sql.Named("status", p.Status),
		sql.Named("address", p.Address),
		sql.Named("created_At", p.CreatedAt))
	if err != nil {
		return 0, err
	}

	id, _ := res.LastInsertId()

	return int(id), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	row := s.db.QueryRow("SELECT number, client, status, address, created_At FROM parcel WHERE number = :number",
		sql.Named("number", number))

	// заполните объект Parcel данными из таблицы
	p := Parcel{}

	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		// возвращаем частично заполненный Parsel или создать новый?
		return Parcel{}, err
	}

	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	rows, err := s.db.Query("SELECT number, client, status, address, created_At FROM parcel WHERE client = :client",
		sql.Named("client", client))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// заполните срез Parcel данными из таблицы
	var res []Parcel

	for rows.Next() {
		p := Parcel{}
		err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			if err == sql.ErrNoRows {
				continue // Если правильно понял - если по Client мы не нашли посылок то это не ошибка
			}
			return nil, err
		}
		res = append(res, p)
	}
	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	// реализуйте обновление статуса в таблице parcel
	res, err := s.db.Exec("UPDATE parcel SET status = :status WHERE number = :number",
		sql.Named("status", status),
		sql.Named("number", number))

	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}

	// не совсем понял, вернётся ли ошибка из Exec если он не повлияет ни на 1 строку, но лучше проверить:
	if count == 0 {
		return fmt.Errorf("не удалось выставить статус %s для посылки %d", status, number)
	}

	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// Есть 2 стула: делать SELECT для просмотра статуса записи в БД || делать UPDATE с жестким условием по статусу
	// + меньше запросов к бд
	// + между запросом статуса и выполнением UPDATE кто-то может статус изменить
	// - нет явной ошибки в случае невыполнения UPDATE (запись не изменилась потому что невалидный ID || потому что статус не тот?)

	res, err := s.db.Exec("UPDATE parcel SET address = :address WHERE number = :number AND status = :status",
		sql.Named("address", address),
		sql.Named("number", number),
		sql.Named("status", ParcelStatusRegistered))

	// Аналогично, 	не совсем понял, вернётся ли ошибка из Exec если он не повлияет ни на 1 строку - в данном случае может быть критично.
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf("не удалось сменить адрес для посылки %d", number)
	}

	return nil
}

func (s ParcelStore) Delete(number int) error {
	// реализуйте удаление строки из таблицы parcel
	// удалять строку можно только если значение статуса registered

	res, err := s.db.Exec("DELETE FROM parcel WHERE number = :number AND status = :status",
		sql.Named("number", number),
		sql.Named("status", ParcelStatusRegistered))

	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf("не удалось удалить посылку %d", number)
	}

	return nil
}
