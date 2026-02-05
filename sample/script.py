import requests
import subprocess

# This function fetches data
def fetch_data():
    response = requests.get('https://api.github.com')
    return response.json()

# Execute a system command
def run_script():
    subprocess.run(["ls", "-l"])
