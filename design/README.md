## Overview
* Cibus offers cupons for different shopping vendors (e.g., Shufersal, Be, Victory)
* It is possible to use the Cibus allowance for buying such cupons
* In such case, the buyer is delivered the cupon details by EMail and SMS
* Currently, there is no built-in way in Cibus or the vendors to keep track of the cupons (details, status, etc.)
* This bot is developed to help in the task of organizing and keeping track of Cibus cupons

## Requirements
* Add a cupon by posting a picture of it
* List cupons
* Use cupons (mark as used)

## Solution highlight
### Telegram bot mechanism
* Bot is registered in Telegram (via BotFather)
* Web server implementing bot's buisness logic is hosted in one of the cloud providers
* Web server is registered in Telegram as a webhook for bot updates 
* Requests to the bot are passed by Telegram to the web server for processing

### Scheme
* Cupon
  * id (string)
  * index (int)
  * vendor (string)
  * value (float)
  * expiration (date)
  * used (bool)
  * used_date (date)

### Data store
* Cupons will be saved in 2 json files (new & used) in an object store of one of the cloud providers
  
### Commands
* Add cupon
  * Implicit command (i.e. not invoked by sending a "/\<cmd\>" message)
  * Done by sending a picture containing cupon details (id, vendor, value, expiration date)
  * Data is extracted using OCR helper utils
  * Cupon object is created and saved
  
* List cupons
  * /list
  * List all new cupons sorted by expiration and value
  
* Use cupons
  * /use a1 a2 ... an
  * Mark cupons with indices a1,a2..an as used
  
### Auth / Privacy
  * Bot is publicly available in Telegram
  * For keeping the cupons safe, the bot will ask for a password from unauthenticated users before processing any command
  * Upon sucessfull password, user will be considered authenticated and allowed to invoke commands
  * Bot will respond to messages only on private chats (i.e. not groups/channels)
