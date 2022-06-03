## Overview
* Cibus offers coupons for different shopping vendors (e.g., Shufersal, Be, Victory)
* It is possible to use the Cibus allowance for buying such coupons
* In such case, the buyer is delivered the coupon details by EMail and SMS
* Currently, there is no built-in way in Cibus or the vendors to keep track of the coupons (details, status, etc.)
* This bot is developed to help in the task of organizing and keeping track of Cibus coupons

## Requirements
* Ingest new coupons automatically
* List coupons
* Use coupons (mark as used)

## Solution highlight

### Schema
* Coupon
  * id (string)
  * vendor (string)
  * value (int)
  * expiration (date)

### Data store
* Coupons will be saved in an object store (e.g., s3)

### Coupons ingestion
* New coupon orders arrive by mail
* A forward rule is forwarding the mails to a special "mailbox" managed by Amazon SES
* SES keeps the mails in the coupons bucket in a special folder (e.g., /email)
* SES triggers a special lambda (i.e., ingester) that process the mail and extracts the coupon attributes
* Ingester stores the coupon in the bucket (e.g., in /new)
<img width="1563" alt="image" src="https://user-images.githubusercontent.com/12379320/171339394-3798e93e-be4e-43d4-8606-8747007c2cd5.png">


### Telegram bot mechanism
* Bot is registered in Telegram (via BotFather)
* Web server implementing bot's business logic is hosted in one of the cloud providers
* Web server is registered in Telegram as a webhook for bot updates 
* Requests to the bot are passed by Telegram to the web server for processing
  
### Commands
* List coupons
  * /list
  * List all new coupons
  <img width="994" alt="image" src="https://user-images.githubusercontent.com/12379320/171339998-5523164b-c36f-463a-9b21-979c2024f397.png">
  
* Use coupons
  * /use c1
  * Mark coupon with id c1 as used
  <img width="1011" alt="image" src="https://user-images.githubusercontent.com/12379320/171340131-a685b10f-aa13-4a2f-9f36-047840eb7197.png">
  
### Auth / Privacy
  * Bot is publicly available in Telegram
  * For keeping the coupons safe, the bot will only process commands originated from allowed users
  * List of allowed users will be injected to the bot's BE (e.g., as ENV var)
  * Bot will respond to messages only on private chats (i.e. not groups/channels)
