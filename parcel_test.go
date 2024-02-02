package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)

	store = getStore()
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func getStore() ParcelStore {
	db, err := sql.Open("sqlite", "tracker.db")
	// require.NoError(t, err, "Database connection should be")
	if err != nil {
		fmt.Println(err)
		panic("No db connection")
	}
	// require.NotEmpty(t, db, "DB object should not be empty")

	return NewParcelStore(db)
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err, "Insert operation should not come to error")
	require.Greaterf(t, id, 0, "Indentifier should return number that more that 0, but got %d", id)
	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	var p Parcel

	p, err = store.Get(id)
	require.NoError(t, err, "Get operation should not come to error")
	require.Greaterf(t, p.Number, 0, "Parcel should have number field greater or equal 0, but has %d", p.Number)

	require.Equal(t, parcel.Address, p.Address, "Values should be the same")
	require.Equal(t, parcel.Client, p.Client, "Values should be the same")
	require.Equal(t, parcel.Status, p.Status, "Values should be the same")
	require.Equal(t, parcel.CreatedAt, p.CreatedAt, "Values should be the same")

	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	err = store.Delete(id)
	require.NoError(t, err, "Delete operation should not come to error")

	p, err = store.Get(id)
	require.Error(t, err, "Get operation should not come to error")
	require.Empty(t, p, "Delete functon should remove parcel from db but it did not do it")

}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err, "Insert operation should not comes to error")
	require.Greaterf(t, id, 0, "Indentifier should return number that more that 0, but got %d", id)

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err, "Operation should not comes to error")

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	var p Parcel
	p, err = store.Get(id)
	require.NoError(t, err, "Operation should not comes to error")
	require.Equal(t, newAddress, p.Address, "Address should be updated")
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err, "Insert operation should not come to error")
	require.Greaterf(t, id, 0, "Indentifier should return number that more that 0, but got %d", id)

	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err, "Update operation should not come to error")
	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	var p Parcel
	p, err = store.Get(id)
	require.NoError(t, err, "Get operation should not come to error")
	require.Equal(t, p.Status, ParcelStatusSent, "Status should be updated")
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i]) // добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
		require.NoError(t, err, "Insert operation should not come to error")
		require.Greaterf(t, id, 0, "Indentifier should return number that more that 0, but got %d", id)
		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client) // получите список посылок по идентификатору клиента, сохранённого в переменной client
	// убедитесь в отсутствии ошибки
	require.NoError(t, err, "Get operation should not come to error")
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных
	require.Equal(t, len(storedParcels), len(parcels), "Parcels count should be the same")

	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		p := parcelMap[parcel.Number]
		require.Equal(t, parcel, p, "Should be the same")
		// убедитесь, что значения полей полученных посылок заполнены верно
		require.Equal(t, parcel.Address, p.Address, "Values should be the same")
		require.Equal(t, parcel.Client, p.Client, "Values should be the same")
		require.Equal(t, parcel.Status, p.Status, "Values should be the same")
		require.Equal(t, parcel.CreatedAt, p.CreatedAt, "Values should be the same")
	}

	defer store.db.Close()
}
