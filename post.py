class Post:
    def __init__(self, pid, plikenum, pbadge, pcontent, ptime, pquote=None):
        self.id = pid
        self.likenum = plikenum
        self.badge = pbadge
        self.content = pcontent
        self.time = ptime
        self.quote = pquote
        self.replies = []

    def add_reply(self, rid, name, rcontent, rtime, rquote):
        self.replies.append(Reply(rid, name, rcontent, rtime, rquote))


class Reply:
    def __init__(self, rid, name, rcontent, rtime, rquote):
        self.id = rid
        self.name = name
        self.content = rcontent
        self.time = rtime
        self.quote = rquote