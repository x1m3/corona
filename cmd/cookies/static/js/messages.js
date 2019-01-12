const ViewPortRequestType = 1;
const ViewPortResponseType = 2;
const UserJoinRequestType = 3;
const UserJoinResponseType = 4;
const CreateCookieRequestType = 5;
const CreateCookieResponseType= 6;



function ViewPortRequest(x, y, xx, yy, angle, turbo) {
    this.t = ViewPortRequestType;
    this.d = {X:x, Y:y, XX:xx, YY:yy, R:angle, T:turbo}
};

function UserJoinRequest(username) {
    this.t = UserJoinRequestType;
    this.d = {UN:username};
};

function UserJoinResponse(ok, altNames) {
    this.t = UserJoinResponseType;
    this.ok = ok;
    this.altNames = altNames;
}

function CreateCookieRequest() {
    this.t = CreateCookieRequestType;
    this.d = null;
}