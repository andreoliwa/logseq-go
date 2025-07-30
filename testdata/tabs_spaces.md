- preserve tabs and spaces in the beginning of empty lines
- _if removed, they will be added back by Logseq_
- level 1
	- level 2
		- Level 3 with tabs and spaces in the following line
		  
		  ![:bulb:](https://a.slack-edge.com/123.png) **Some text with tabs and spaces in the beginning of the line.**
- preserve tabs and spaces in code blocks
	- ```
	  ❯ poetry install
	  Creating virtualenv whatsapp-parser-lite-master-E_oyzMpG-py3.9 in /Users/zzz/Library/Caches/pypoetry/virtualenvs
	  Updating dependencies
	  Resolving dependencies... (3.4s)
	  
	  Writing lock file
	  
	  Package operations: 5 installs, 0 updates, 0 removals
	  
	    • Installing six (1.16.0)
	    • Installing numpy (1.24.2)
	    • Installing python-dateutil (2.8.2)
	    • Installing pytz (2022.7.1)
	    • Installing pandas (1.5.3)
	  
	  /Users/zzz/Downloads/whatsapp-parser-lite-master/whatsapp_parser_lite_master does not contain any element
	  
	  ❯ pp parse_whatsapp.py ../WhatsApp-chat.txt output.csv
	  
	  # 3 empty lines at the end of the block
	  
	  
	  
	  ```
