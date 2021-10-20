import axios, {AxiosInstance} from 'axios';

export const FormatRaw = "raw"

export interface ViewSpec {
    Name: string
    Env: EnvSpec[]
    Steps: Step[]
}

export interface Step {
    Name: string
    Env: EnvSpec[]
    Command: string
}

export interface EnvSpec {
    Name: string
}

export interface EnvValue {
    Name: string
    Value: string
}

export interface StepOutput {
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

export interface GetStepOutputRequest {
    ViewName: string
    ViewEnv?: EnvValue[]
    StepName: string
    StepEnv?: EnvValue[]
    Format?: string
}

export interface GetStepOutputResponse extends ErrorResponse {
    Output: StepOutput
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

    GetStepOutputLink(req: GetStepOutputRequest): string {
        let url = new URL(this.opts.config.addr + "/getStepOutput")
        url.searchParams.append("step_name", req.StepName)
        url.searchParams.append("view_name", req.ViewName)
        if (req.Format && req.Format !== "") {
            url.searchParams.append("format", req.Format)
        }
        if (req.ViewEnv) {
            req.ViewEnv.forEach((v: EnvValue) => {
                url.searchParams.append("view_env" + v.Name, v.Value)
            })
        }
        if (req.StepEnv) {
            req.StepEnv.forEach((v: EnvValue) => {
                url.searchParams.append("step_env" + v.Name, v.Value)
            })
        }

        return url.toString()
    }

    async GetViewOuput(req: GetStepOutputRequest): Promise<GetStepOutputResponse> {
        let rsp = await this.client.request<GetStepOutputResponse>({
            url: this.GetStepOutputLink(req),
            method: "get",
            data: req,
            headers: {
                "Content-Type": "application/json; charset=utf-8",
            },
        });

        return rsp.data
    }
}