from flask import Flask
import pika
import uuid
import requests
from requests.auth import HTTPBasicAuth

app = Flask(__name__)

connection = pika.BlockingConnection(pika.ConnectionParameters("localhost"))
channel = connection.channel()


@app.route("/weather/<location>")
def weather(location: str):
    corr_id = str(uuid.uuid4())
    channel.basic_publish(
        exchange="weather",
        routing_key="",
        properties=pika.BasicProperties(
            correlation_id=corr_id,
        ),
        body=location,
    )
    return corr_id


@app.route("/weather/results/<uuid>")
def weather_results(uuid: str):
    method_frame, header_frame, body = channel.basic_get(queue=uuid, auto_ack=True)
    return body


@app.route("/metrics")
def metrics():
    auth = HTTPBasicAuth("guest", "guest")
    response = requests.get("http://localhost:15672/api/overview", auth=auth)
    return response.json()
