from locust import HttpLocust, TaskSet, task

class WebsiteTasks(TaskSet):
    
    @task
    def index(self):
        self.client.get("/community/id/540dae3162c7744fbb000004")

class WebsiteUser(HttpLocust):
    host = "http://localhost:8080/http"
    task_set = WebsiteTasks
    min_wait = 5000
    max_wait = 15000
