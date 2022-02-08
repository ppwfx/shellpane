import axios, {AxiosInstance} from 'axios';

export const FormatRaw = "raw"

export interface ViewConfig {
    Name: string
    Execute: ViewExecuteConfig
    Command?: CommandConfig
    Sequence?: SequenceConfig
    Category: CategoryConfig
}

export interface ViewExecuteConfig {
    Auto: boolean
}

export interface CategoryConfig {
    Slug: string
    Name: string
    Color: string
}

export interface SequenceConfig {
    Slug: string
    Steps: StepConfig[]
}

export interface StepConfig {
    Name: string
    Command: CommandConfig
}

export interface CommandConfig {
    Slug: string
    Command: string
    Display: string
    Description: string
    Inputs: CommandInputConfig[]
}

export interface CommandInputConfig {
    Name: string
    Input: InputConfig
}

export interface InputConfig {
    Slug: string
    Description: string
}

export interface ResponseError {
    Code: string
    Message: string
}

export interface ErrorResponse {
    Error: ResponseError
}

export interface ExecuteCommandRequest {
    Slug: string
    Inputs: InputValue[]
    Format?: string
}

export interface InputValue {
    Name: string
    Value: string
}

export interface ExecuteCommandResponse extends ErrorResponse {
    Output: CommandOutput
}

export interface CommandOutput {
    Stdout: string
    Stderr: string
    ExitCode: number
}

export interface GetViewConfigsRequest {
}

export interface GetViewConfigsResponse extends ErrorResponse {
    ViewConfigs?: ViewConfig[]
}

export interface GetCategoryConfigsRequest {
}

export interface GetCategoryConfigsResponse extends ErrorResponse {
    CategoryConfigs?: CategoryConfig[]
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

    async GetViewConfigs(req: GetViewConfigsRequest): Promise<GetViewConfigsResponse> {
        let rsp = await this.client.request<GetViewConfigsResponse>({
            url: this.opts.config.addr + "/getViewConfigs",
            method: "get",
            data: req,
            headers: {
                "Content-Type": "application/json; charset=utf-8",
            },
        });

        return rsp.data
    }

    async GetCategoryConfigs(req: GetCategoryConfigsRequest): Promise<GetCategoryConfigsResponse> {
        let rsp = await this.client.request<GetCategoryConfigsResponse>({
            url: this.opts.config.addr + "/getCategoryConfigs",
            method: "get",
            data: req,
            headers: {
                "Content-Type": "application/json; charset=utf-8",
            },
        });

        return rsp.data
    }

    ExecuteCommandLink(req: ExecuteCommandRequest): string {
        let url = new URL(this.opts.config.addr + "/executeCommand")
        url.searchParams.append("slug", req.Slug)
        if (req.Format && req.Format !== "") {
            url.searchParams.append("format", req.Format)
        }
        if (req.Inputs) {
            req.Inputs.forEach((v: InputValue) => {
                url.searchParams.append("input_" + v.Name, v.Value)
            })
        }

        return url.toString()
    }

    async ExecuteCommand(req: ExecuteCommandRequest): Promise<ExecuteCommandResponse> {
        let rsp = await this.client.request<ExecuteCommandResponse>({
            url: this.ExecuteCommandLink(req),
            method: "get",
            data: req,
            headers: {
                "Content-Type": "application/json; charset=utf-8",
            },
        });

        return rsp.data
    }
}