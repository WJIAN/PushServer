

return redis.call('HMSET', KEYS[1],
                  'remote', ARGV[1],
                  'appid', ARGV[2],
                  'installid', ARGV[3],
                  'restaddr', ARGV[4]
)
