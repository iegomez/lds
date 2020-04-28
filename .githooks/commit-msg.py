import sys
import re

with open(sys.argv[1],mode='r') as file:
    msg = file.read()
    if re.search(r'(fixes|closes|refs) #\d+', msg):
    	sys.exit(0)
    else:
        print("Bad commit message {}".format(msg))
        sys.exit(1)

