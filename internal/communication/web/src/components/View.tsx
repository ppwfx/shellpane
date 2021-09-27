import React, {useEffect} from 'react';
import * as client from '../client';
import {EnvSpec} from '../client';
import {message} from 'antd';

interface ViewProps {
    client: client.Client
    viewSpec: client.ViewSpec
}

export const View = (props: ViewProps) => {
    const [viewOutputRsp, setViewOutputRsp] = React.useState<client.GetViewOutputResponse | undefined>(undefined);
    const [viewOutputReq, setViewOutputReq] = React.useState<client.GetViewOutputRequest>({
        Name: props.viewSpec.Name,
        Env: [],
    });
    const [isLoading, setIsLoading] = React.useState<boolean>(true);

    const [count, setCount] = React.useState<number>(0);

    let rawReq: client.GetViewOutputRequest = Object.assign({}, viewOutputReq);
    rawReq.Format = client.FormatRaw;

    const setEnvValue = ((name: string, value: string) => {
        let viewOutputReqCopy = Object.assign({}, viewOutputReq);
        if (!viewOutputReqCopy.Env) {
            viewOutputReqCopy.Env = []
        }

        for (let i = 0; i < viewOutputReqCopy.Env?.length; i++) {
            if (viewOutputReqCopy.Env[i].Name === name) {
                viewOutputReqCopy.Env[i].Value = value
                setViewOutputReq(viewOutputReqCopy)

                return
            }
        }

        viewOutputReqCopy.Env.push({Name: name, Value : value})
        setViewOutputReq(viewOutputReqCopy)
    })

    const refresh = () => {setCount(count+1)}

    useEffect(() => {
        (async function () {
            if (props.viewSpec.Env && props.viewSpec.Env.length !== 0 && viewOutputReq.Env?.length === 0) {
                setIsLoading(false);
                return
            }

            setIsLoading(true);

            const viewOutputRsp = await props.client.GetViewOuput(viewOutputReq)

            setViewOutputRsp(viewOutputRsp);
            setIsLoading(false);
        })().catch((reason: any) => {
            message.error('failed to get view specs: ' + reason);
            setViewOutputRsp(undefined);
            setIsLoading(false);
        });
    }, [count]);

    return (
        <div className="views__view">
            <div className="views__view__header">
                <span className="views__view__header__name">
                    {props.viewSpec.Name}
                </span>
                {isLoading ? <span className="views__view__header__loader"><div className="loader01"/></span>:null}
                <span className="views__view__header__raw">
                    <a rel="noreferrer" onClick={refresh}>Run </a>
                    <a href={props.client.GetViewOutputLink(rawReq)} target="_blank" rel="noreferrer">Raw </a>
                    <a href={props.client.GetViewOutputLink(rawReq)} target="_blank" rel="noreferrer" download="file.txt">Download </a>
                </span>
                {props.viewSpec.Env ? <ViewEnv env={props.viewSpec.Env} setEnvValue={setEnvValue} refresh={refresh}/> : null}
            </div>
            <div className="views__view__command">
                <textarea rows={1} value={"$ " + props.viewSpec.Command} disabled={true}/>

            </div>
            <div className="views__view__output">

                <textarea rows={1} value={viewOutputRsp?.Output?.Stdout ? viewOutputRsp?.Output?.Stdout: viewOutputRsp?.Output?.Stderr} disabled={true}/>

                {/*{viewOutputRsp?.Output.Stdout}<br/>*/}
                {/*{viewOutputRsp?.Output.Stderr}<br/>*/}
                {/*exit code: {viewOutputRsp?.Output.ExitCode}*/}

            </div>
        </div>
    );
}

interface ViewEnvProps {
    env: client.EnvSpec[]
    setEnvValue: (k: string, v: string) => void
    refresh: () => void
}

const ViewEnv = (props: ViewEnvProps) => {
    return (
        <span className="view__env">
            {props.env.map((s: EnvSpec) => {
                return <label key={s.Name}>{s.Name}
                    <input type="text" className="view__env__input"
                           onChange={e => props.setEnvValue(s.Name, e.target.value)}
                           onKeyDown={(e) => {
                               if (e.code !== "Enter") {
                                   return
                               }

                               e.preventDefault()

                               props.refresh()
                           }}
                    />
                </label>
            })}
        </span>
    );
}


export default View;
