# DevArch Flask Template

Flask framework template configured to serve static files from `public/` directory.

## Quick Start

```bash
# Install Flask
pip install flask

# Run development server
python app.py
```

## app.py Configuration

```python
from flask import Flask

# Configure Flask to use public/ for static files
app = Flask(__name__,
            static_folder='public',
            static_url_path='')

@app.route('/')
def index():
    return app.send_static_file('index.html')

@app.route('/api/hello')
def hello():
    return {'message': 'Hello from Flask API'}

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8300, debug=True)
```

## Structure

```
flask-app/
├── public/              # WEB ROOT
│   ├── index.html
│   └── assets/
├── app.py              # Main application
├── requirements.txt
└── README.md
```

## DevArch Integration

### Port: 8300-8399 (Python range)
### Domain: Configure via Nginx Proxy Manager
### Container: python service

## Documentation

See: https://flask.palletsprojects.com/
