
import {createTwirpRequest, throwTwirpError, Fetch} from './twirp';


export interface Application {
    name?: string;
}

interface ApplicationJSON {
    name?: string;
}



const JSONToApplication = (m: Application | ApplicationJSON): Application => {
    if (m === null) {
		return null;
	}
    return {
        name: m.name,
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

