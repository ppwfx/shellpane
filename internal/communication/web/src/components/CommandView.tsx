import React, {useEffect, useRef} from 'react';
import * as client from '../client';
import {InputValue} from '../client';
import {message, Popover} from 'antd';

interface CommandViewProps {
    client: client.Client
    name: string
    viewConfig: client.ViewConfig
    commandConfig: client.CommandConfig
}

export const CommandView = (props: CommandViewProps) => {
    const [executeCommandRsp, setExecuteCommandRsp] = React.useState<client.ExecuteCommandResponse | undefined>(undefined);
    const [inputValues, setInputValues] = React.useState<InputValue[]>([]);
    const [isLoading, setIsLoading] = React.useState<boolean>(false);
    const [count, setCount] = React.useState<number>(0);
    const [updateCount, setUpdateCount] = React.useState<number>(0);

    const firstInputRef = React.createRef();

    let stepsBodyRef = useRef<any>(null);

    const setInputValue = ((name: string, value: string) => {
        let inputValuesCopy = [...inputValues]
        for (let i = 0; i < inputValuesCopy?.length; i++) {
            if (inputValuesCopy[i].Name === name) {
                inputValuesCopy[i].Value = value
                setInputValues(inputValuesCopy)

                return
            }
        }

        inputValuesCopy.push({Name: name, Value: value})
        setInputValues(inputValuesCopy)
    })

    const refresh = () => {
        setCount(count + 1)
    }

    useEffect(() => {
        if (!props.viewConfig.Execute.Auto && count === 0) {
            return
        }

        let inputConfigs = props.commandConfig.Inputs

        if (inputConfigs && inputConfigs.length !== 0 && inputValues?.length === 0) {
            setIsLoading(false);

            return
        }

        setIsLoading(true);

        (async function () {
            let req: client.ExecuteCommandRequest = {
                Slug: props.commandConfig.Slug,
                Inputs: inputValues,
            }
            const executeCommandRsp = await props.client.ExecuteCommand(req)

            setExecuteCommandRsp(executeCommandRsp);
            setIsLoading(false);
            setUpdateCount(updateCount + 1)
        })().catch((reason: any) => {
            message.error('failed to get step output: ' + reason);
            setIsLoading(false);
        });
    }, [count]);

    const filename = `${props.name.replaceAll(" ", "_")}_${(new Date()).toISOString().slice(0, 19).replace("T", "_")}.txt`


    let className = "views__view--active"

    let rawReq: client.ExecuteCommandRequest = {
        Slug: props.commandConfig.Slug,
        Inputs: inputValues,
    }
    rawReq.Format = client.FormatRaw;

    let outputClassName = "views__view__output"

    let viewClassName = `views__view a-color--${props.viewConfig.Category.Slug} input-background--${props.viewConfig.Category.Slug}`

    // @ts-ignore
    let body = <div className={className} key={props.name}>
        <div className="views__view__step__header">
            {isLoading
            ? <span className="views__view__header__loader">
                <div className={"loader--" + props.viewConfig.Category.Slug}/>
            </span>
            : null}

            {props.commandConfig.Inputs ?
                <ViewEnv inputConfigs={props.commandConfig.Inputs}
                         inputValues={inputValues}
                         ref={firstInputRef}
                         setInputValue={(k, v: string) => setInputValue(k, v)}
                         disabled={isLoading}
                    // disabled={false}
                         refresh={refresh}/> : null}
        </div>
        <div
            className={outputClassName}>{executeCommandRsp?.Output?.Stdout ? executeCommandRsp?.Output?.Stdout : executeCommandRsp?.Output?.Stderr}</div>
    </div>

    return (
        <div className={viewClassName} key={props.name}>
            <div className="views__view__header">
                <span className="views__view__header__name">
                    {props.name}
                </span>
                <span className="views__view__header__raw">
                    <a rel="noreferrer" onClick={refresh}>Run </a>
                    <a href={props.client.ExecuteCommandLink(rawReq)} target="_blank"
                       rel="noreferrer">Raw </a>
                    <a href={props.client.ExecuteCommandLink(rawReq)} target="_blank" rel="noreferrer"
                       download={filename}>Download </a>
                </span>
            </div>
            <div className="views__view__body" ref={stepsBodyRef}>
                <div className="views__view__steps">
                    {body}
                </div>
            </div>
        </div>
    );
}

interface ViewEnvProps {
    inputConfigs: client.CommandInputConfig[]
    inputValues: client.InputValue[]
    setInputValue: (k: string, v: string) => void
    disabled: boolean
    refresh: () => void
}

const ViewEnv = React.forwardRef((props: ViewEnvProps, ref: any) => {
    let values: any = {};
    props.inputValues.forEach(v => {
        values[v.Name] = v.Value;
    });

    return (
        <span className="view__env">
            {props.inputConfigs.map((s: client.CommandInputConfig, i: number) => {
                return <label key={s.Input.Slug}>{s.Input.Slug}
                    {s.Input.Description ? <Popover content={s.Input.Description} trigger="hover">?</Popover> : null}
                    <input type="text" className="view__env__input"
                           value={values[s.Input.Slug] ? values[s.Input.Slug] : ""}
                           disabled={props.disabled}
                           ref={i == 0 && ref ? ref : null}
                           onChange={e => props.setInputValue(s.Input.Slug, e.target.value)}
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
})


export default CommandView;
