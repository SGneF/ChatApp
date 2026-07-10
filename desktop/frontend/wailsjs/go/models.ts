export namespace main {
	
	export class ApplyFriendRequest {
	    to_user_id: number;
	    remark: string;
	
	    static createFrom(source: any = {}) {
	        return new ApplyFriendRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.to_user_id = source["to_user_id"];
	        this.remark = source["remark"];
	    }
	}
	export class BackendStatus {
	    ok: boolean;
	    url: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new BackendStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.url = source["url"];
	        this.message = source["message"];
	    }
	}
	export class FriendRequestResponse {
	    id: number;
	    from_user_id: number;
	    from_username: string;
	    from_nickname: string;
	    from_avatar: string;
	    remark: string;
	    status: string;
	    create_time: string;
	
	    static createFrom(source: any = {}) {
	        return new FriendRequestResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.from_user_id = source["from_user_id"];
	        this.from_username = source["from_username"];
	        this.from_nickname = source["from_nickname"];
	        this.from_avatar = source["from_avatar"];
	        this.remark = source["remark"];
	        this.status = source["status"];
	        this.create_time = source["create_time"];
	    }
	}
	export class FriendResponse {
	    id: number;
	    username: string;
	    nickname: string;
	    avatar: string;
	    signature: string;
	    remark: string;
	
	    static createFrom(source: any = {}) {
	        return new FriendResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.username = source["username"];
	        this.nickname = source["nickname"];
	        this.avatar = source["avatar"];
	        this.signature = source["signature"];
	        this.remark = source["remark"];
	    }
	}
	export class HandleFriendRequest {
	    request_id: number;
	
	    static createFrom(source: any = {}) {
	        return new HandleFriendRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.request_id = source["request_id"];
	    }
	}
	export class LoginRequest {
	    username: string;
	    password: string;
	
	    static createFrom(source: any = {}) {
	        return new LoginRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.username = source["username"];
	        this.password = source["password"];
	    }
	}
	export class UserResponse {
	    id: number;
	    username: string;
	    nickname: string;
	    avatar: string;
	    signature: string;
	    create_time: string;
	    update_time: string;
	
	    static createFrom(source: any = {}) {
	        return new UserResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.username = source["username"];
	        this.nickname = source["nickname"];
	        this.avatar = source["avatar"];
	        this.signature = source["signature"];
	        this.create_time = source["create_time"];
	        this.update_time = source["update_time"];
	    }
	}
	export class LoginResponse {
	    token: string;
	    user: UserResponse;
	
	    static createFrom(source: any = {}) {
	        return new LoginResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.token = source["token"];
	        this.user = this.convertValues(source["user"], UserResponse);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class RegisterRequest {
	    username: string;
	    password: string;
	    nickname: string;
	
	    static createFrom(source: any = {}) {
	        return new RegisterRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.username = source["username"];
	        this.password = source["password"];
	        this.nickname = source["nickname"];
	    }
	}
	export class SearchUserResponse {
	    id: number;
	    username: string;
	    nickname: string;
	    avatar: string;
	    signature: string;
	
	    static createFrom(source: any = {}) {
	        return new SearchUserResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.username = source["username"];
	        this.nickname = source["nickname"];
	        this.avatar = source["avatar"];
	        this.signature = source["signature"];
	    }
	}
	export class UpdateProfileRequest {
	    nickname: string;
	    avatar: string;
	    signature: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateProfileRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nickname = source["nickname"];
	        this.avatar = source["avatar"];
	        this.signature = source["signature"];
	    }
	}

}

