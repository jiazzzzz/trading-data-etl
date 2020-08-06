import sys
import datetime

class Logger:
    '''
    Add Logger for cases and libs.

    Example: logger = Logger("test-s3",1)

            logger.info("Found some issue")
    '''
    def __init__(self,log_module,log_level=0):
        self.log_module = log_module
        #self.logger = open(log_file,"a")
        self.log_level = log_level

    def _write_log(self,log_type,log_str):
        _now_time = datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')
        _log = "%s [%s] [%s] %s"%(_now_time,log_type,self.log_module,log_str)
        #self.logger.write("%s\n"%(_log))
        print(_log)        

    def info(self,log_str):
        self._write_log("INFO",log_str)
    
    def debug(self,log_str):
        if self.log_level>0:
            self._write_log("DEBUG",log_str)
    
    def warn(self,log_str):
        self._write_log("WARN",log_str)
    
    def err(self,log_str):
        self._write_log("ERROR",log_str)

    def critical(self,log_str):
        self._write_log("CRITICAL",log_str)

if __name__ == '__main__':
    logger = Logger("test-s3",1)
    logger.info("Found some issue")
    logger.debug("Try to fix it.")


