
import {createTwirpRequest, throwTwirpError, Fetch} from './twirp';


export interface LoginReq {
    state?: string;
}

interface LoginReqJSON {
    state?: string;
}



const LoginReqToJSON = (m: LoginReq): LoginReqJSON => {
	if (m === null) {
		return null;
	}
	
    return {
        state: m.state,
    };
};


export interface LoginRes {
    redirectUrl?: string;
}

interface LoginResJSON {
    redirect_url?: string;
}



const JSONToLoginRes = (m: LoginRes | LoginResJSON): LoginRes => {
    if (m === null) {
		return null;
	}
    return {
        redirectUrl: (((m as LoginRes).redirectUrl) ? (m as LoginRes).redirectUrl : (m as LoginResJSON).redirect_url),
    };
};


export interface Application {
    name?: string;
    type?: string;
}

interface ApplicationJSON {
    name?: string;
    type?: string;
}



const JSONToApplication = (m: Application | ApplicationJSON): Application => {
    if (m === null) {
		return null;
	}
    return {
        name: m.name,
        type: m.type,
    };
};


export interface AddApplicationReq {
    owner?: string;
    name?: string;
    url?: string;
    path?: string;
    branch?: string;
    deploymentType?: string;
    privateKey?: string;
    dryRun?: boolean;
    private?: boolean;
    namespace?: string;
    dir?: string;
}

interface AddApplicationReqJSON {
    owner?: string;
    name?: string;
    url?: string;
    path?: string;
    branch?: string;
    deployment_type?: string;
    private_key?: string;
    dry_run?: boolean;
    private?: boolean;
    namespace?: string;
    dir?: string;
}



const AddApplicationReqToJSON = (m: AddApplicationReq): AddApplicationReqJSON => {
	if (m === null) {
		return null;
	}
	
    return {
        owner: m.owner,
        name: m.name,
        url: m.url,
        path: m.path,
        branch: m.branch,
        deployment_type: m.deploymentType,
        private_key: m.privateKey,
        dry_run: m.dryRun,
        private: m.private,
        namespace: m.namespace,
        dir: m.dir,
    };
};


export interface AddApplicationRes {
    application?: Application;
}

interface AddApplicationResJSON {
    application?: ApplicationJSON;
}



const JSONToAddApplicationRes = (m: AddApplicationRes | AddApplicationResJSON): AddApplicationRes => {
    if (m === null) {
		return null;
	}
    return {
        application: JSONToApplication(m.application),
    };
};


export interface GitOps {
    login: (loginReq: LoginReq) => Promise<LoginRes>;
    
    addApplication: (addApplicationReq: AddApplicationReq) => Promise<AddApplicationRes>;
    
}

export class DefaultGitOps implements GitOps {
    private hostname: string;
    private fetch: Fetch;
    private writeCamelCase: boolean;
    private pathPrefix = "/twirp/gitops.GitOps/";
    private headersOverride: HeadersInit;

    constructor(hostname: string, fetch: Fetch, writeCamelCase = false, headersOverride: HeadersInit = {}) {
        this.hostname = hostname;
        this.fetch = fetch;
        this.writeCamelCase = writeCamelCase;
        this.headersOverride = headersOverride;
    }
    login(loginReq: LoginReq): Promise<LoginRes> {
        const url = this.hostname + this.pathPrefix + "Login";
        let body: LoginReq | LoginReqJSON = loginReq;
        if (!this.writeCamelCase) {
            body = LoginReqToJSON(loginReq);
        }
        return this.fetch(createTwirpRequest(url, body, this.headersOverride)).then((resp) => {
            if (!resp.ok) {
                return throwTwirpError(resp);
            }

            return resp.json().then(JSONToLoginRes);
        });
    }
    
    addApplication(addApplicationReq: AddApplicationReq): Promise<AddApplicationRes> {
        const url = this.hostname + this.pathPrefix + "AddApplication";
        let body: AddApplicationReq | AddApplicationReqJSON = addApplicationReq;
        if (!this.writeCamelCase) {
            body = AddApplicationReqToJSON(addApplicationReq);
        }
        return this.fetch(createTwirpRequest(url, body, this.headersOverride)).then((resp) => {
            if (!resp.ok) {
                return throwTwirpError(resp);
            }

            return resp.json().then(JSONToAddApplicationRes);
        });
    }
    
}

