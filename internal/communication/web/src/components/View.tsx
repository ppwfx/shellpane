import React, {useEffect, useRef} from 'react';
import * as client from '../client';
import {EnvSpec, EnvValue} from '../client';
import {message} from 'antd';

interface ViewProps {
    client: client.Client
    viewSpec: client.ViewSpec
}

export const View = (props: ViewProps) => {
    const [currentExecIndex, setCurrentExecIndex] = React.useState<number>(0);
    const [stepOutputRsps, setStepOutputRsps] = React.useState<client.GetStepOutputResponse[]>([...Array(props.viewSpec.Steps.length)]);
    const [viewEnv, setViewEnv] = React.useState<EnvValue[]>([]);
    const [stepEnvs, setStepEnvs] = React.useState<EnvValue[][]>([...Array.from({length: props.viewSpec.Steps.length}, (v, i) => [])]);
    const [isLoading, setIsLoading] = React.useState<boolean>(true);
    const [count, setCount] = React.useState<number>(0);

    const stepsCount = props.viewSpec.Steps.length
    const hasMultipleSteps = stepsCount > 1
    const hasViewExec = props.viewSpec.Env?.length != 0
    const isViewExec = hasViewExec && currentExecIndex == 0
    const execCount = props.viewSpec.Steps.length + (hasViewExec ? 1 : 0)
    let currentStepIndex = hasViewExec ? currentExecIndex === 0 ? undefined : currentExecIndex - 1 : currentExecIndex

    const firstInputRefs = useRef([]);
    // @ts-ignore
    firstInputRefs.current = [0, 0, 0, 0, 0].map((ref, index) => {
        return React.createRef()
    })

    let stepsBodyRef = useRef<any>(null);

    console.log(currentExecIndex, currentStepIndex)

    const setViewEnvValue = ((name: string, value: string) => {
        let viewEnvCopy = [...viewEnv]
        for (let i = 0; i < viewEnvCopy?.length; i++) {
            if (viewEnvCopy[i].Name === name) {
                viewEnvCopy[i].Value = value
                setViewEnv(viewEnvCopy)

                return
            }
        }

        viewEnvCopy.push({Name: name, Value: value})
        setViewEnv(viewEnvCopy)
    })

    const setStepEnvValue = ((i: number, name: string, value: string) => {
        let stepEnvsCopy = [...stepEnvs]
        for (let ii = 0; ii < stepEnvsCopy[i]?.length; ii++) {
            if (stepEnvsCopy[i][ii].Name === name) {
                stepEnvsCopy[i][ii].Value = value
                setStepEnvs(stepEnvsCopy)

                return
            }
        }

        stepEnvsCopy[i].push({Name: name, Value: value})
        setStepEnvs(stepEnvsCopy)
    })

    const refresh = () => {
        setCount(count + 1)
    }

    const handleFocus = (execIndex: number) => {
        setTimeout(() => {
            let blurIndex = execIndex === 0 ? execCount - 1 : execIndex - 1
            // @ts-ignore
            if (firstInputRefs.current[blurIndex].current != null) {
                // @ts-ignore
                firstInputRefs.current[blurIndex].current.blur()
            }

            // @ts-ignore
            if (firstInputRefs.current[execIndex].current != null) {
                // @ts-ignore
                firstInputRefs.current[execIndex].current.focus()
            }
        }, 100)
    }

    useEffect(() => {
        let envSpec: client.EnvSpec[] | undefined = undefined
        let envValues: client.EnvValue[] | undefined = undefined

        if (isViewExec) {
            envSpec = props.viewSpec.Env
            envValues = viewEnv
        } else {
            if (currentStepIndex === undefined) {
                alert("currentStepIndex === undefined")
                return
            }

            envSpec = props.viewSpec.Steps[currentStepIndex].Env
            envValues = stepEnvs[currentStepIndex]
        }

        if (envSpec && envSpec.length !== 0 && envValues?.length === 0) {
            setIsLoading(false);

            return
        }

        let currentExecIndexCopy = currentExecIndex
        if (isViewExec) {
            currentExecIndexCopy++
            setCurrentExecIndex(currentExecIndexCopy)
            setStepOutputRsps([])
            setStepEnvs([...Array.from({length: stepsCount}, (v, i) => [])]);

            handleFocus(currentExecIndexCopy)

            return
        }

        if (currentStepIndex === undefined) {
            alert("currentStepIndex === undefined")
            return
        }

        setIsLoading(true);

        (async function () {
            let req: client.GetStepOutputRequest = {
                ViewName: props.viewSpec.Name,
                ViewEnv: viewEnv,
                StepName: props.viewSpec.Steps[currentStepIndex].Name,
                StepEnv: stepEnvs[currentStepIndex],
            }
            const viewOutputRsp = await props.client.GetViewOuput(req)

            let stepOutputRspsCopy = [...stepOutputRsps]
            stepOutputRspsCopy[currentStepIndex] = viewOutputRsp

            if (currentExecIndex + 1 >= execCount) {
                currentExecIndexCopy = 0
                setViewEnv([]);
            } else {
                currentExecIndexCopy++
                currentStepIndex = hasViewExec ? currentExecIndexCopy === 0 ? undefined : currentExecIndexCopy - 1 : currentExecIndexCopy
            }

            setCurrentExecIndex(currentExecIndexCopy)
            setStepOutputRsps(stepOutputRspsCopy);
            setIsLoading(false);

            if (hasMultipleSteps) {
                stepsBodyRef.current?.scrollTo({
                    top: stepsBodyRef.current.scrollHeight,
                    behavior: 'smooth'
                });
            }

            handleFocus(currentExecIndexCopy)

            if (currentStepIndex && (!props.viewSpec.Steps[currentStepIndex].Env || props.viewSpec.Steps[currentStepIndex].Env.length === 0)) {
                refresh()
            }
        })().catch((reason: any) => {
            message.error('failed to get step output: ' + reason);
            setIsLoading(false);
        });
    }, [count]);

    const filename = `${props.viewSpec.Name.replaceAll(" ", "_")}_${(new Date()).toISOString().slice(0, 19).replace("T", "_")}.txt`

    let className = "views__view__header"
    if (hasViewExec && currentExecIndex === 0) {
        className = className + " views__view__header--active"
    }

    return (
        <div className="views__view" key={props.viewSpec.Name}>
            <div className={className}>
                <span className="views__view__header__name">
                    {props.viewSpec.Name}
                </span>
                <span className="views__view__header__raw">
                    <a rel="noreferrer" onClick={refresh}>Start </a>
                </span>
                {props.viewSpec.Env ?
                    <ViewEnv ref={props.viewSpec.Env.length != 0 ? firstInputRefs.current[0] : null}
                             env={props.viewSpec.Env}
                             envValues={viewEnv} setEnvValue={setViewEnvValue}
                             refresh={refresh}
                             disabled={currentExecIndex !== 0 || isLoading}/> : null}
            </div>
            <div className="views__view__body" ref={stepsBodyRef}>
                <div className="views__view__steps">
                    {props.viewSpec.Steps.map((step: client.Step, stepIndex: number) => {
                        let className = "views__view__step"
                        if (stepIndex === currentStepIndex && !isLoading) {
                            className = className + " views__view__step--active"
                        }

                        let execIndex: number;
                        if (hasViewExec) {
                            execIndex = stepIndex + 1
                        } else {
                            execIndex = stepIndex
                        }

                        let rawReq: client.GetStepOutputRequest = {
                            ViewName: props.viewSpec.Name,
                            ViewEnv: viewEnv,
                            StepName: props.viewSpec.Steps[stepIndex].Name,
                            StepEnv: stepEnvs[stepIndex],
                        }
                        rawReq.Format = client.FormatRaw;

                        let outputClassName = "views__view__output"
                        if (currentStepIndex && stepIndex === currentStepIndex-1 && stepOutputRsps[stepIndex]) {
                            outputClassName = outputClassName + " views__view__output--highlight"
                        }

                        return (
                            <div className={className} key={props.viewSpec.Name + step.Name}>
                                <div className="views__view__step__header">
                                <span className="views__view__step__header__name">
                                    {step.Name}
                                </span>
                                    {isLoading && stepIndex === currentStepIndex ?
                                        <span className="views__view__header__loader"><div
                                            className="loader01"/></span> : null}
                                    <span className="views__view__header__raw">
                                    <a rel="noreferrer" onClick={refresh}>Run </a>
                                    <a href={props.client.GetStepOutputLink(rawReq)} target="_blank"
                                       rel="noreferrer">Raw </a>
                                    <a href={props.client.GetStepOutputLink(rawReq)} target="_blank" rel="noreferrer"
                                       download={filename}>Download </a>
                                </span>
                                    {step.Env ?
                                        <ViewEnv env={step.Env} envValues={stepEnvs[stepIndex]}
                                                 ref={firstInputRefs.current[execIndex]}
                                                 setEnvValue={(k, v: string) => setStepEnvValue(stepIndex, k, v)}
                                                 disabled={execIndex !== currentExecIndex || isLoading}
                                            // disabled={false}
                                                 refresh={refresh}/> : null}
                                </div>
                                <div
                                    className={outputClassName}>{stepOutputRsps[stepIndex]?.Output?.Stdout ? stepOutputRsps[stepIndex]?.Output?.Stdout : stepOutputRsps[stepIndex]?.Output?.Stderr}</div>
                            </div>
                        )
                    })}
                </div>
            </div>
        </div>
    );
}

interface ViewEnvProps {
    env: client.EnvSpec[]
    envValues: client.EnvValue[]
    setEnvValue: (k: string, v: string) => void
    disabled: boolean
    refresh: () => void
}

const ViewEnv = React.forwardRef((props: ViewEnvProps, ref: any) => {
    let values: any = {};
    props.envValues.forEach(v => {
        values[v.Name] = v.Value;
    });

    return (
        <span className="view__env">
            {props.env.map((s: EnvSpec, i: number) => {
                return <label key={s.Name}>{s.Name}
                    <input type="text" className="view__env__input" value={values[s.Name] ? values[s.Name] : ""}
                           disabled={props.disabled}
                           ref={i == 0 && ref ? ref : null}
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
})


export default View;
