import requests

APP_ID = "ab3033ce"
APP_KEY = "b582dd99bf360556355c5ca503cd5526"

URL = "https://api.adzuna.com/v1/api/jobs/in/search/1"

params = {
    "app_id": APP_ID,
    "app_key": APP_KEY,
    "results_per_page": 20,
    "what": "software engineer",
    "where": "india"
}

res = requests.get(URL, params=params)
data = res.json()

for job in data["results"]:
    print({
        "title": job["title"],
        "company": job["company"]["display_name"],
        "location": job["location"]["display_name"],
        "description": job["description"][:200],
        "apply_url": job["redirect_url"]
    })
