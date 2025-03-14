from flask import Flask, Response
import pika
import json
import uuid
import requests
from requests.auth import HTTPBasicAuth

app = Flask(__name__)

connection = pika.BlockingConnection(pika.ConnectionParameters("rabbitmq"))
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
    if body is None:
        return Response("Page not found", status=404)
    return json.loads(body)


@app.route("/metrics")
def metrics():
    auth = HTTPBasicAuth("guest", "guest")
    response = requests.get("http://rabbitmq:15672/api/overview", auth=auth)
    return response.json()


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5000)
