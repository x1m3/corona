const ViewPortRequestType = 1;
const ViewPortResponseType = 2;
const UserJoinRequestType = 3;

function ViewPortRequest(x, y, xx, yy) {
    this.t = ViewPortRequestType;
    this.d = {X:x, Y:y, XX:xx, YY:yy}
}