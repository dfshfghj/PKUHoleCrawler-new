import json
import os

class WebConfig:
    _default_config = {
        "browser": "chrome",
        "profiles_path": "/path/to/profiles",
        "timeout": 30,
        "crawl_size": 100000
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
    def timeout(self):
        return self._config.get("timeout", self._default_config["timeout"])

    @timeout.setter
    def timeout(self, value):
        self._config["timeout"] = value
        self.save_config()

    @property
    def crawl_size(self):
        return self._config.get("crawl_size", self._default_config["crawl_size"])

    @crawl_size.setter
    def crawl_size(self, value):
        self._config["crawl_size"] = value
        self.save_config()


if __name__ == "__main__":
    config = WebConfig()


    print("Current Browser:", config.browser)
    print("Current Profiles Path:", config.profiles_path)
    print("Current Timeout:", config.timeout)


    config.browser = "Firefox"
    config.profiles_path = "D:\\PKUHoleCrawler-master\\Profiles\\32fy5laa.default-release"
    config.timeout = 60
    config.crawl_size = 10000


    print("Updated Browser:", config.browser)
    print("Updated Profiles Path:", config.profiles_path)
    print("Updated Timeout:", config.timeout)