
local cid = KEYS[1]
local msgid = ARGV[1]
local pb = ARGV[2]
local stamp = ARGV[3]

local mk = 'M.'..string.sub(cid, 1, 5)..'.'..msgid
--local mk = 'M.'..cid..'.'..msgid
local smk = 'SM.'..cid

redis.call('SET', mk, pb)

-- 超过一个小时消息不发了
redis.call('EXPIRE', mk, 3600)


local num = redis.call('ZCARD', smk)

redis.call('ZREMRANGEBYRANK', smk, 100, -1)

redis.call('ZADD', smk,
           stamp,
           msgid
)

-- 3600*24*7 = 604800
redis.call('EXPIRE', smk, 604800)
