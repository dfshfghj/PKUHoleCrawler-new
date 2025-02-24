import json
import os
import argparse

class WebConfig:
    _default_config = {
        "browser": "chrome",
        "profiles_path": "/path/to/profiles",
        "crawl_size": 1000,
        "part": 200,
        "mode": "Normal"
    }

    _config_file = "config.json"

    def __init__(self):
        self.load_config()

    def load_config(self):
        if os.path.exists(self._config_file):
            with open(self._config_file, 'r') as f:
                self._config = json.load(f)
        else:
            self._config = self._default_config
            self.save_config()

    def save_config(self):
        with open(self._config_file, 'w') as f:
            json.dump(self._config, f, indent=4)

    @property
    def browser(self):
        return self._config.get("browser", self._default_config["browser"])

    @browser.setter
    def browser(self, value):
        self._config["browser"] = value
        self.save_config()

    @property
    def profiles_path(self):
        return self._config.get("profiles_path", self._default_config["profiles_path"])

    @profiles_path.setter
    def profiles_path(self, value):
        self._config["profiles_path"] = value
        self.save_config()

    @property
    def crawl_size(self):
        return self._config.get("crawl_size", self._default_config["crawl_size"])

    @crawl_size.setter
    def crawl_size(self, value):
        self._config["crawl_size"] = value
        self.save_config()

    @property
    def part(self):
        return self._config.get("part", self._default_config["part"])

    @part.setter
    def part(self, value):
        self._config["part"] = value
        self.save_config()

    @property
    def mode(self):
        return self._config.get("mode", self._default_config["mode"])

    @mode.setter
    def mode(self, value):
        self._config["mode"] = value
        self.save_config()

if __name__ == "__main__":
    parse = argparse.ArgumentParser()

    parse.add_argument('--mode', choices=['Normal', 'Full', 'Specific'])
    parse.add_argument('--crawl_size', type=int, default=1000)
    parse.add_argument('--part', type=int, default=200)
    parse.add_argument('--browser',choices=['Firefox', 'Edge'])
    parse.add_argument('--profiles_path')

    args = vars(parse.parse_args())

    webconfig = WebConfig()
    webconfig.browser = args['browser']
    webconfig.profiles_path = args['profiles_path']
    webconfig.crawl_size = args['crawl_size']
    webconfig.part = args['part']

    