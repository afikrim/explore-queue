from locust import HttpUser, constant_pacing, task


class Ping(HttpUser):
    wait_time = constant_pacing(1)

    @task
    def ping(self):
        self.client.get(url="/ping", timeout=60)
