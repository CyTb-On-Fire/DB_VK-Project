build:
	docker build -t park .

run:
	docker run -rm -p 5000:5000 park

