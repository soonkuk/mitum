import uuidv4 from 'uuid/v4'

class Time {
  constructor() {
    this.t = null
    this.n = null
  }

  static parse(s) {
    var iso = /(\d{4})-([01]\d)-([0-3]\d)T([0-2]\d):([0-5]\d):([0-5]\d)\.(\d+)([+-][0-2]\d:[0-5]\d|Z)/

    var p = s.match(iso);
    if (p == null) {
      throw new Error('invalid time', s)
    }

    var t = new Date(
      p[1],
      Number.parseInt(p[2], 10) - 1,
      p[3],
      p[4],
      p[5],
      p[6],
      Number.parseInt(p[7].slice(0, 3), 10),
    )

    var time = new Time()
    time.t = t
    time.n = (t.getTime() * 1000) + Number.parseInt(p[7].slice(0, 6), 10)

    return time
  }

  elapsed(t) {
    var d = this.n - t.n
    if (d < 0) {
      return '000.000000s'
    }

    var s = Math.floor(d / 1000000)
    var head = s.toString()
    if (head.length < 3) {
      head = '0'.repeat(3 - head.length) + head
    }

    var tail = (d- (s * 1000000)).toString()
    if (tail.length < 6) {
      tail = '0'.repeat(6 - tail.length) + tail
    }

    return head + '.' + tail + 's'
  }
}

class Record {
  constructor() {
    this.module = null
    this.message = null
    this.level = null
    this.t = null
    this.node = null
    this.caller = null
    this.extra = null
    this.body = null
    this.id = null
  }

  static fromJSONString(line) {
    var o = JSON.parse(line)

    // module
    var module = null
    if ('module' in o === false) {
      throw new Error('module is missing')
    } else {
      module = o['module']
    }

    // message
    var message = null
    if ('msg' in o === false) {
      throw new Error('message is missing')
    } else {
      message = o['msg']
    }

    // level
    var level = null
    if ('lvl' in o === false) {
      throw new Error('level is missing')
    } else {
      level = o['lvl']
    }

    // t
    var t = null
    var id = null
    if ('t' in o === false) {
      throw new Error('t is missing')
    } else {
      t = Time.parse(o['t'])
      id = t.n + '-' + uuidv4()
    }

    // node
    var node = null
    if ('node' in o === false) {
      //
    } else {
      node = o['node']
    }

    // caller
    var caller = null
    if ('caller' in o === false) {
      throw new Error('caller is missing')
    } else {
      caller = o['caller']
    }

    // extra
    var extra = null;
    delete o.msg
    delete o.lvl
    delete o.t
    delete o.caller
    delete o.node
    delete o.module

    extra = o;

    // body
    var r = new Record()

    r.id = id
    r.t = t
    r.module = module
    r.message = message
    r.level = level
    r.node = node
    r.caller = caller
    r.extra = extra
    r.body = line;

    return r
  }
}

class Log {
  constructor() {
    this.nodes = []
    this.records = []
  }

  static load (contents) {
    var log = new Log()

    var nodes = []
    var line = ''
    for (const c of contents) {
      if (c === '\n') {
        var record = null
        record = this.parseRecord(line)
        line = ''

        if (record === undefined) {
          continue
        }

        log.records.push(record)

        // node
        if (record.node == null) {
          continue
        } else if (nodes.includes(record.node)) {
          continue
        } else {
          nodes.push(record.node)
        }

        continue
      }

      line += c
    }

    nodes.sort()
    log.nodes = nodes

    return log
  }

  static parseRecord(line) {
    var record = Record.fromJSONString(line)
    if (record.node == null) {
      return
    }

    return record
  }
}

export default Log
//module.exports = Log