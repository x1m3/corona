function JSONMarshalUnmarshal() {
    this.encode = function(obj) {
        return JSON.stringify(obj);
    };
    this.decode = function(buffer) {
        return JSON.parse(buffer);
    };
}

function MsgPackMarshalUnmarshal() {
    this.encode = function(obj) {
        return msgpack.encode(obj);
    };
    this.decode = function(buffer) {
        return msgpack.decode(buffer);
    };

}

function Transport(wsUrl, coder) {
    _this = this;
    this.coder = coder;
    this.conn = new WebSocket(wsUrl);
    this.callbacks = new Map();

    this.conn.onopen = function () {
        console.log("socket is open")
    };

    this.conn.onmessage = function (e) {
        rawMsg = _this.coder.decode(e.data);
        fn = _this.callbacks.get(rawMsg.t);
        fn(rawMsg);
    };

    this.conn.onerror = function (e) {
        console.log("we have a network error");
    };

    this.send = function(msg) {
        if (_this.conn.readyState === _this.conn.OPEN) {
            _this.conn.send(_this.coder.encode(msg));
        }
    };

    this.registerCallback = function(msgType, fn) {
        _this.callbacks.set(msgType, fn)
    }
 }