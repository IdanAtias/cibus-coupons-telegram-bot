## Overview
* Cibus offers coupons for different shopping vendors (e.g., Shufersal, Be, Victory)
* It is possible to use the Cibus allowance for buying such coupons
* In such case, the buyer is delivered the coupon details by EMail and SMS
* Currently, there is no built-in way in Cibus or the vendors to keep track of the coupons (details, status, etc.)
* This bot is developed to help in the task of organizing and keeping track of Cibus coupons

## Requirements
* Add a coupon by posting a picture of it
* List coupons
* Use coupons (mark as used)

## Solution highlight
### Telegram bot mechanism
* Bot is registered in Telegram (via BotFather)
* Web server implementing bot's business logic is hosted in one of the cloud providers
* Web server is registered in Telegram as a webhook for bot updates 
* Requests to the bot are passed by Telegram to the web server for processing

### Schema
* Coupon
  * id (string)
  * index (int)
  * vendor (string)
  * value (float)
  * expiration (date)
  * used (bool)
  * used_date (date)

### Data store
* coupons will be saved in 2 json files (new & used) in an object store of one of the cloud providers
  
### Commands
* Add coupon
  * Implicit command (i.e. not invoked by sending a "/\<cmd\>" message)
  * Done by sending a picture containing coupon details (id, vendor, value, expiration date)
  * Data is extracted using OCR helper utils
  * coupon object is created and saved
  
* List coupons
  * /list
  * List all new coupons sorted by expiration and value
  
* Use coupons
  * /use a1 a2 ... an
  * Mark coupons with indices a1,a2..an as used
  
### Auth / Privacy
  * Bot is publicly available in Telegram
  * For keeping the coupons safe, the bot will ask for a password from unauthenticated users before processing any command
  * Upon successful password, user will be considered authenticated and allowed to invoke commands
  * Bot will respond to message only on private chats (i.e. not groups/channels)
