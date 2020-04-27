import sys
import re
from pathlib import Path

msg = Path(sys.argv[1]).read_text()
if re.search(r'(fixes|closes|refs) #\d+', msg):
	sys.exit(0)
else:
	print("Bad commit message ", msg)
	sys.exit(1)

