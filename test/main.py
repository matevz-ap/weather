from flask import Flask, Response
import pika
import json
import uuid
import requests
from pika.adapters.blocking_connection import BlockingChannel
from requests.auth import HTTPBasicAuth

app = Flask(__name__)

connection: pika.BlockingConnection = pika.BlockingConnection(
    pika.ConnectionParameters("rabbitmq")
)
channel: BlockingChannel = connection.channel()


@app.route("/weather/<location>")
def weather(location: str) -> str:
    """
    Adds message to RabbitMQ to be processed.
    :parama location: location name ex. Ljubljana
    :returns: uuid to be used to get results
    """
    corr_id: str = str(uuid.uuid4())
    channel.basic_publish(
        exchange="weather",
        routing_key="",
        properties=pika.BasicProperties(
            correlation_id=corr_id,
        ),
        body=location.encode(),
    )
    return corr_id


@app.route("/weather/stress/<amount>")
def wrather_stress(amount: str) -> dict:
    uuids: list[str] = []
    for _ in range(int(amount)):
        corr_id: str = str(uuid.uuid4())
        channel.basic_publish(
            exchange="weather",
            routing_key="",
            properties=pika.BasicProperties(
                correlation_id=corr_id,
            ),
            body="Ljubljana".encode(),
        )
        uuids.append(corr_id)
    return {"uuids": uuids, "count": len(uuids)}


@app.route("/weather/results/<uuid>")
def weather_results(uuid: str) -> Response | dict:
    """
    Endpoint checks if the reuqest with uuid was processed
    and returns the results in case it was.
    :param uuid: uuid returned from /weather endpoint to identify the result
    :returns: Weather data for location or 404 of data no present in RabbitMQ
    """
    _, _, body = channel.basic_get(queue=uuid, auto_ack=True)
    if body is None:
        return Response("Page not found", status=404)
    return json.loads(body)


@app.route("/metrics")
def metrics() -> dict:
    auth = HTTPBasicAuth("guest", "guest")
    response: requests.Response = requests.get(
        "http://rabbitmq:15672/api/overview", auth=auth
    )
    return response.json()


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5000)
