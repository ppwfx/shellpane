import axios, {AxiosInstance} from 'axios';

export const FormatRaw = "raw"

export interface ViewSpec {
    Name: string
    Description: string
    Command: string
    Env: EnvSpec[]
}

export interface EnvSpec {
    Name: string
}

export interface EnvValue {
    Name: string
    Value: string
}

export interface ViewOutput {
    Stdout: string
    Stderr: string
    ExitCode: number
}

export interface ResponseError {
    Code: string
    Message: string
}

export interface ErrorResponse {
    Error: ResponseError
}

export interface GetViewOutputRequest {
    Name: string
    Format?: string
    Env?: EnvValue[]
}

export interface GetViewOutputResponse extends ErrorResponse {
    Output: ViewOutput
}

export interface GetViewSpecsRequest {
}

export interface GetViewSpecsResponse extends ErrorResponse {
    Specs?: ViewSpec[]
}

export interface ClientConfig {
    addr: string;
}

export interface ClientOpts {
    config: ClientConfig;
}

export class Client {
    opts: ClientOpts;
    client: AxiosInstance;

    constructor(opts: ClientOpts) {
        this.opts = opts;

        this.client = axios.create({})
    }

    async GetViewSpecs(req: GetViewSpecsRequest): Promise<GetViewSpecsResponse> {
        let rsp = await this.client.request<GetViewSpecsResponse>({
            url: this.opts.config.addr + "/getViewSpecs",
            method: "get",
            data: req,
            headers: {
                "Content-Type": "application/json; charset=utf-8",
            },
        });

        return rsp.data
    }

    GetViewOutputLink(req: GetViewOutputRequest): string {
        let url = new URL(this.opts.config.addr + "/getViewOutput")
        url.searchParams.append("name", req.Name)
        if (req.Format && req.Format !== "") {
            url.searchParams.append("format", req.Format)
        }
        if (req.Env) {
            req.Env.forEach((v: EnvValue)=> {
                url.searchParams.append("env" + v.Name, v.Value)
            })
        }

        return url.toString()
    }

    async GetViewOuput(req: GetViewOutputRequest): Promise<GetViewOutputResponse> {
        let rsp = await this.client.request<GetViewOutputResponse>({
            url: this.GetViewOutputLink(req),
            method: "get",
            data: req,
            headers: {
                "Content-Type": "application/json; charset=utf-8",
            },
        });

        return rsp.data
    }
}