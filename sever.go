package main

import (
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
)

type Basket struct {
	ID     int
	Values []string
}

var baskets []Basket
var dataAmount = 0

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/init", initHandler)
	http.HandleFunc("/add-basket", addBasketHandler)
	http.HandleFunc("/add-data", addDataHandler)
	http.HandleFunc("/get-data", getDataHandler)

	http.ListenAndServe(":8080", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// Отображение текущих данных о корзинах
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("<h1>Baskets</h1>"))
	w.Write([]byte("<div id=\"basket-container\">"))
	w.Write([]byte("</div>"))
	w.Write([]byte(`
		<button onclick="addBasket()">Add Basket</button>
		<button onclick="addData()">Add Data</button>
		<button onclick="getData()">Get Data</button>
		<div id="logging"></div>
		<style>
			.highlight  { background-color:yellow; }
		</style>
		<script>
			function getBasketElement(basketNumber, values) {
				const basketDiv = document.createElement('div');
				basketDiv.class = 'basket';
				basketDiv.setAttribute('data-basketid', basketNumber);
				basketDiv.innerHTML = '<strong>Basket ' +  basketNumber + ' (<span class="len">' + values.length +'</span> items):</strong><span class="elems">[' + values.join(', ') + ']</span>';
				return basketDiv;
			}
			function addLog(text) {
				let loggingDiv = document.getElementById("logging");
				var newParagraph = document.createElement("p");
				newParagraph.innerHTML = text;

				// Вставляем новый элемент в начало <div>
				loggingDiv.insertBefore(newParagraph, loggingDiv.firstChild);
			}
			function addBasket() {
				fetch('/add-basket', { method: 'POST' })
					.then(response => response.json())
					.then(data => {
						if (data.success) {
							const newBasketID = data.newBasketID;
							const basketContainer = document.getElementById('basket-container');
							const newBasketDiv = getBasketElement(newBasketID, []);
							basketContainer.appendChild(newBasketDiv);
							addLog("New basket added: " + newBasketID);
						}
					})
					.catch(error => console.error('Error:', error));
			}

			function addData() {

				fetch('/add-data', {
					method: 'POST',
					headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
				})
					.then(response => response.json())
					.then(data => {
						if (data.success) {
							const newDataID = data.dataID;
							const basketID = data.basketID;
							const basketDiv = document.querySelector('div[data-basketid="' + basketID + '"]');
							basketDiv.querySelector('.len').innerText = Number.parseInt(basketDiv.querySelector('.len').innerText) + 1;
							currentText = basketDiv.querySelector('.elems').innerText.slice(0, -1);
							basketDiv.querySelector('.elems').innerText = currentText + newDataID + ', ]';
							addLog("New data " + newDataID +" added to busket " + basketID)
						}
					})
					.catch(error => console.error('Error:', error));
			}

			function getData() {
				const dataID = prompt('Enter the data ID:');
				if (!dataID) return;

				fetch('/get-data', {
					method: 'POST',
					headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
					body: 'fileId=' + dataID,
				})
					.then(response => response.json())
					.then(data => {
						if (data.success) {
							const basketID = data.basketID;
							const basketDiv = document.querySelector('div[data-basketid="' + basketID + '"]');
			
							// Добавляем класс для подсветки на две секунды
							basketDiv.classList.add('highlight');
							setTimeout(() => {
								// Удаляем класс подсветки после двух секунд
								basketDiv.classList.remove('highlight');
							}, 2000);
							addLog("Data " + dataID +" found in busket " + basketID)
						} else {
							alert('Error: ' + data.error);
						}
					})
					.catch(error => console.error('Error:', error));
			}

			function initBaskets() {
				fetch('/init', { method: 'POST' })
					.then(response => response.json())
					.then(data => {
						if (data.success) {
							const basketsContainer = document.getElementById('basket-container');
							basketsContainer.innerHTML = ''; // Очистка контейнера перед добавлением новых корзин
	
							const baskets = data.baskets || [];
							amount = data.amount || 0;
							baskets.forEach(basket => {
								const basketElement = getBasketElement(basket.ID, basket.Values);
								basketsContainer.appendChild(basketElement);
							});
						}
					})
					.catch(error => console.error('Error:', error));
			}
	
			// Инициализация при загрузке страницы
			initBaskets();

		</script>
	`))
}

func initHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Initialized.")
	// Инициализация - создание начальных данных о корзинах
	initBasketsState()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "baskets": baskets})
}

func addBasketHandler(w http.ResponseWriter, r *http.Request) {

	// Добавление новой корзины
	newBasketID, err := addBasket()
	if err != nil {
		log.Println("Cannot add basket.")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Cannot add data"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "newBasketID": newBasketID})
}

func addDataHandler(w http.ResponseWriter, r *http.Request) {
	// Добавление данных в корзину
	dataID := dataAmount + 1

	basketID, err := addData(dataID)
	if err != nil {
		log.Println("Cannot add data.")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Cannot add data"})
		return
	}
	log.Printf("Added Data %d to Basket %d.", dataID, basketID)
	dataAmount++

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "dataID": dataID, "basketID": basketID})
}

func getDataHandler(w http.ResponseWriter, r *http.Request) {

	// Считывание данных из тела запроса
	err := r.ParseForm()
	if err != nil {
		log.Printf("Error parsing form: %s", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Error parsing form"})
		return
	}

	// Получение dataId из формы
	dataID := r.FormValue("dataId")
	if dataID == "" {
		log.Printf("Error: dataId is empty")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "dataId is empty"})
		return
	}

	// Поиск случайного бакета
	basketID, err := saveToBasket()
	if err != nil {
		log.Println("Not found basket of data %s.", dataID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Not found"})
		return
	}
	log.Printf("Requested Data %s from Basket %d.", dataID, basketID)

	// Возвращаем номер бакета
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "basketID": basketID})

}

func join(values []string) string {
	return "[" + strconv.Quote(strings.Join(values, ", ")) + "]"
}

///////////////////// Функции алгоритма /////////////////////

func saveToBasket() (int, error) {
	if len(baskets) > 1 {
		return rand.Intn(len(baskets)) + 1, nil
	}
	return 0, errors.New("Баскет не найден")
}

func addBasket() (int, error) {
	newBasketID := len(baskets) + 1
	log.Printf("Added Basket %d.", newBasketID)
	baskets = append(baskets, Basket{ID: newBasketID, Values: []string{}})
	return newBasketID, nil
}

func addData(dataID int) (int, error) {
	// Добавление данных в случайную корзину
	basketID := rand.Intn(len(baskets)) + 1
	baskets[basketID-1].Values = append(baskets[basketID-1].Values, strconv.Itoa(dataID))
	return basketID, nil
}

func initBasketsState() {
	if len(baskets) == 0 {
		log.Println("Added Basket 1.")
		baskets = append(baskets, Basket{ID: 1, Values: []string{}})
	}
}
