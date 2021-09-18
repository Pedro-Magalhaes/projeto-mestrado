import sys
import time
from watchdog.observers import Observer
from watchdog.events import LoggingEventHandler
import logging

class Filewatcher:
    def __init__(self) -> None:
        logging.basicConfig(level=logging.INFO,
                    format='%(asctime)s - %(message)s',
                    datefmt='%Y-%m-%d %H:%M:%S')
        self.event_handler = LoggingEventHandler()
        self.observers = {}
    
    def start(self, path = '.'): # FIXME: hash the observers só we can add more!
        pathhash = hash(path)
        if pathhash not in self.observers or self.observers[pathhash] == None:
            observer = Observer()
            observer.schedule(self.event_handler, path, recursive=False)
            observer.start()
            self.observers[pathhash] = observer

    def stop(self, path = '.'):
        pathhash = hash(path)
        if pathhash in self.observers and self.observers[pathhash] != None:
            observer = self.observers[pathhash]
            observer.stop()
            observer.join()
            self.observers[pathhash] = None



# if __name__ == "__main__":
    # logging.basicConfig(level=logging.INFO,
    #                     format='%(asctime)s - %(message)s',
    #                     datefmt='%Y-%m-%d %H:%M:%S')
#     logging.log()
#     path = sys.argv[1] if len(sys.argv) > 1 else '.'
#     event_handler = LoggingEventHandler()
#     observer = Observer()
#     observer.schedule(event_handler, path, recursive=True)
#     observer.start()
#     try:
#         while True:
#             time.sleep(1)
#     finally:
#         observer.stop()
#         observer.join()