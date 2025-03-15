# weather

Run `docker compose up --build` in root directory to start containters.

The app consists of RabbitMQ service, go consumer, flask REST service and Valkey service for caching.

## Flask

Flask REST service implements 4 endpoints:

### 1. /weather/<location_name>
Providing location (ex. Ljubljana), returns uuid that can be used to get the results later.

### 2. /weather/stress/<amount>
Provide amount off messages you want to emit to queue. Receive a list of result ids

### 2. /weather/results/<uuid>
Providing an uuid, returns wether results from location.

### 3. /metrics
Returns basic metrics from RabbitMQ