from logger import Logger
import configparser

class Common():
    def __init__(self):
        pass

    def read_conf(self, config_file, section, key):
        config = configparser.ConfigParser()
        config.read(config_file)
        value = config.get(section, key)
        return value

if __name__ == '__main__':
    t = Common()
    v = t.read_conf("settings.conf", 'db', 'passwd')
    print(v)