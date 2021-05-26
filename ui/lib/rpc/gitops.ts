
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


export interface ListApplicationsReq {
}

interface ListApplicationsReqJSON {
}


const ListApplicationsReqToJSON = (_: ListApplicationsReq): ListApplicationsReqJSON => {
    return {};
};


export interface ListApplicationsRes {
    applications?: Application[];
}

interface ListApplicationsResJSON {
    applications?: ApplicationJSON[];
}



const JSONToListApplicationsRes = (m: ListApplicationsRes | ListApplicationsResJSON): ListApplicationsRes => {
    if (m === null) {
		return null;
	}
    return {
        applications: (m.applications as (Application | ApplicationJSON)[]).map(JSONToApplication),
    };
};


export interface GitOps {
    listApplications: (listApplicationsReq: ListApplicationsReq) => Promise<ListApplicationsRes>;
    
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
    listApplications(listApplicationsReq: ListApplicationsReq): Promise<ListApplicationsRes> {
        const url = this.hostname + this.pathPrefix + "ListApplications";
        let body: ListApplicationsReq | ListApplicationsReqJSON = listApplicationsReq;
        if (!this.writeCamelCase) {
            body = ListApplicationsReqToJSON(listApplicationsReq);
        }
        return this.fetch(createTwirpRequest(url, body, this.headersOverride)).then((resp) => {
            if (!resp.ok) {
                return throwTwirpError(resp);
            }

            return resp.json().then(JSONToListApplicationsRes);
        });
    }
    
}

