

local cid = KEYS[1]
local ck = 'I.'..cid

redis.call('HMSET', ck,
                  'restaddr', ARGV[1],
                  'timeout', ARGV[2]+600,

                  'remote', ARGV[3],
                  'appid', ARGV[4],
                  'installid', ARGV[5],
                  'nettype', ARGV[6]
)

-- 3600*24*7 = 604800
return redis.call('EXPIRE', ck, 604800)



