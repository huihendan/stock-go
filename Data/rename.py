#coding:utf-8

import time
import os
import time
import datetime
from fnmatch import fnmatch
import shutil

CAPT_PATH = './'
CAPT_PATH_OUT = '/home/beven/SVN_sockBusiness/Python/snapTest/out_img2/'
CAPT_PATH_TEST = '/home/beven/SVN_sockBusiness/Python/snapTest/'
CAPT_PATH_PRE = '/home/beven/SVN_sockBusiness/Python/snapPre/'
Windows = False
if Windows:
    CAPT_PATH = 'D:\\Project\stockBusinessSVN\Data\Day_k_new'
    CAPT_PATH_OUT = 'D:\\Project/stockBusinessSVN/Python/snapTest/out_img2/'
    CAPT_PATH_TEST = 'D:\\Project/stockBusinessSVN/Python/snap/'
    CAPT_PATH_PRE = 'D:\\Project/stockBusinessSVN/Python/snapPre/'
if not os.path.exists(CAPT_PATH):
    os.mkdir(CAPT_PATH)

def test():
    index = 0
    for file in os.listdir(CAPT_PATH):
        if fnmatch(file, '60*.csv'):
            index +=1
            print(file)
            os.rename(os.path.join(CAPT_PATH,file), os.path.join(CAPT_PATH,"sh." + file))
        if fnmatch(file, '00*.csv') or fnmatch(file, '30*.csv'):
            index +=1
            print(file)
            os.rename(os.path.join(CAPT_PATH,file), os.path.join(CAPT_PATH,"sz."+file))

if __name__ == '__main__':
    test()




