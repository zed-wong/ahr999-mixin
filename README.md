# Ahr999-mixin
A bot that offers up-to-date ahr999 index for subscribed user, and a web page which provides history data chart of the index.



## File explanation

 - main.py 
    	- Run every day to keep data.json up to date.
 - data.json 
	- Json file that contains timestamp and ahr999 index
 - main.go 
     - Execute every 24 hours to update the index
   - Notify with mixin when the index hit the line
   - Handles the bots message module, writes subed userid to database
 - data.db 
    	- Stores subscribed users
