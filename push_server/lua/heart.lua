

local cid = KEYS[1]
local ck = 'I.'..cid

redis.call('HMSET', ck,
                  'remote', ARGV[1],
                  'appid', ARGV[2],
                  'installid', ARGV[3],
                  'restaddr', ARGV[4]
                  'timeout', ARGV[5]+600
)

-- 3600*24*7 = 604800
return redis.call('EXPIRE', ck, 604800)



