import hashlib
import hmac
import time


def get_auth_info(dev_id, dev_key):
    ts = str(int(time.time()))
    id_ts = str(dev_id) + ts
    signature = hmac.new(str(dev_key).encode(), id_ts.encode(),
                         digestmod=hashlib.sha256).hexdigest()
    return [('x-api-key', ','.join([str(dev_id), ts, signature]))]
