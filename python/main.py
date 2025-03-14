from flask import Flask
import pika
import uuid

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
