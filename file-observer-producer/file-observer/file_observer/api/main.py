import uvicorn
from fastapi import FastAPI
from file_observer.watcher.file_watcher import Filewatcher

app = FastAPI()

f = Filewatcher()

@app.get("/")
async def root():
    f.start('/local/ProgramasLocais/Documents/pessoal/Puc/mestrado/projeto-mestrado/file-observer-producer/file-observer')
    return {"message": "Hello World"}

@app.get("/s")
async def root():
    f.stop('/local/ProgramasLocais/Documents/pessoal/Puc/mestrado/projeto-mestrado/file-observer-producer/file-observer')
    return {"message": "Hello World"}

def start():
    """Launched with `poetry run start` at root level"""
    uvicorn.run("file_observer.api.main:app", host="0.0.0.0", port=8000, reload=True)