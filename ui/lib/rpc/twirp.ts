
export interface TwirpErrorJSON {
    code: string;
    msg: string;
    meta: {[index:string]: string};
}

export class TwirpError extends Error {
    code: string;
    meta: {[index:string]: string};

    constructor(te: TwirpErrorJSON) {
        super(te.msg);

        this.code = te.code;
        this.meta = te.meta;
    }
}

export const throwTwirpError = (resp: Response) => {
    return resp.json().then((err: TwirpErrorJSON) => { throw new TwirpError(err); })
};

export const createTwirpRequest = (url: string, body: object, headersOverride: HeadersInit = {}): Request => {
    const headers = {
        ...{
            "Content-Type": "application/json"
        },
        ...headersOverride
    };
    return new Request(url, {
        method: "POST",
        headers,
        body: JSON.stringify(body)
    });
};

export type Fetch = (input: RequestInfo, init?: RequestInit) => Promise<Response>;
