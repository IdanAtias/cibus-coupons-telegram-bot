import json
import email
import re
import base64
import boto3
from datetime import datetime, timezone

BUCKET_NAME = "cibus-coupons-testing"

VENDOR_PHONE_TO_VENDOR_NAME = {
    "04-8247645": "Shufersal-Vardia"
}

def lambda_handler(event, context):
    print("start processing cibus coupon mail")

    # get mail content
    # event holds just a reference (messageId) for the full mail content in s3
    # the messageId is the key name for the mail object in the bucket
    s3 = boto3.client('s3')
    msg_id = event["Records"][0]["ses"]["mail"]["messageId"]
    print(f"msg id: {msg_id}")
    mail_obj = s3.get_object(Bucket=BUCKET_NAME, Key=f"email/{msg_id}")
    mail_str = str(mail_obj["Body"].read().decode("utf-8"))

    # construct the cibus coupon
    # cibus coupon mail has the relevant attributes in the first "text/plain" section
    # track it and extract coupon attributes based on regex matching
    coupon = {}
    msg = email.message_from_string(mail_str)
    for part in msg.walk():
        if part.get_content_type() == "text/plain":
            data = str(base64.b64decode(bytes(part.get_payload(), "utf8")))
            coupon["id"] = re.search("91[\d]+", data).group(0)
            coupon["value"] = int(re.search("[\d]+.00", data).group(0).split(".")[0])  # e.g., '40.00' -> 40
            vendor_phone = re.search("0[\d]+-[\d]+", data).group(0)
            coupon["vendor"] = VENDOR_PHONE_TO_VENDOR_NAME[vendor_phone]
            expiration_str = re.search("[\d]+\/[\d]+\/[\d]+", data).group(0)
            expiration_datetime = datetime.strptime(expiration_str, "%d/%m/%Y")
            expiration_datetime_utc = expiration_datetime.replace(tzinfo=timezone.utc)
            coupon["expiration"] = int(expiration_datetime_utc.timestamp())
            break
    print(f"detected coupon '{coupon}'")

    # validate that coupon doesn't already exist (as new or used)
    cid = coupon["id"]
    for key in [cid, f"used/{cid}"]:
        try:
            s3.get_object(Bucket=BUCKET_NAME, Key=key)
            print(f"coupon already exists at '{BUCKET_NAME}/{key}'. skipping ingestion")
            return {'statusCode': 200}
        except s3.exceptions.NoSuchKey:
            pass

    print(f"saving new coupon to '{BUCKET_NAME}/{cid}'")
    s3.put_object(
        Bucket=BUCKET_NAME,
        Key=cid,
        Body=bytes(json.dumps(coupon), "utf8"),
    )
    return {'statusCode': 200}

