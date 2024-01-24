package main

const clientContent = `
		<button onclick="addBasket()">Add Basket</button>
		<button onclick="addData()">Add Data</button>
		<button onclick="getData()">Get Data</button>
		<div id="logging"></div>
		<style>
			.highlight  { background-color:yellow; }
		</style>
		<script>
			const idsFilePath = 'ids.txt'
			
			function getBasketElement(basketNumber) {
				const basketDiv = document.createElement('div');
				basketDiv.class = 'basket';
				basketDiv.setAttribute('data-basketid', basketNumber);
				basketDiv.innerHTML = '<strong>Basket ' +  basketNumber + ' (<span class="len">' + 0 +'</span> items):</strong>';
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
							const newShardId = data.newShardId;
							const basketContainer = document.getElementById('shard-container');
							const newShardDiv = getBasketElement(newShardId, []);
							basketContainer.appendChild(newShardDiv);
							addLog("New basket added: " + newShardId);
						}
					})
					.catch(error => console.error('Error:', error));
			}

			function addData() {

				fetch('/add-data', {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify({'source_file_name': idsFilePath})
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
							const shardId = data.shardId;
							const basketDiv = document.querySelector('div[data-basketid="' + shardId + '"]');
			
							// Добавляем класс для подсветки на две секунды
							basketDiv.classList.add('highlight');
							setTimeout(() => {
								// Удаляем класс подсветки после двух секунд
								basketDiv.classList.remove('highlight');
							}, 2000);
							addLog("Data " + dataID +" found in shard " + shardId)
						} else {
							alert('Error: ' + data.error);
						}
					})
					.catch(error => console.error('Error:', error));
			}

			function getLogRecord() {
				fetch('/get-log-record', { method: 'GET' })
					.then(response => response.json())
					.then(data => {
						if (data.success) {
							const basketDiv = document.querySelector('div[data-basketid="' + data.shardId + '"]');
							// Добавляем класс для подсветки на две секунды
							basketDiv.classList.add('highlight');
							setTimeout(() => {
								// Удаляем класс подсветки после двух секунд
								basketDiv.classList.remove('highlight');
							}, 1000);
							addLog("File " + data.fileId + " was added to " + data.shardId)
						}
					})
					.catch(error => console.error('Error:', error));
			}

			function initBaskets() {
				fetch('/init', { method: 'POST' })
					.then(response => response.json())
					.then(data => {
						if (data.success) {
							const basketsContainer = document.getElementById('shard-container');
							basketsContainer.innerHTML = ''; // Очистка контейнера перед добавлением новых корзин
	
							amount = data.shardAmount || 0;
							console.log("Shard amount is", amount);
							for (let i = 0; i < amount; i++) {
								const basketElement = getBasketElement(i);
								basketsContainer.appendChild(basketElement);
							}
						}
					})
					.catch(error => console.error('Error:', error));
			}
	
			// Инициализация при загрузке страницы
			initBaskets();

			const t = setInterval(function() { getLogRecord() }, 2000)
			
			setTimeout(() => {
				console.log("Clear interval")
				clearInterval(t)
			}, 60000);
		</script>
	`
