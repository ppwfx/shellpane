import React, {useEffect, useRef} from 'react';
import * as client from '../client';
import {InputValue} from '../client';
import {message} from 'antd';

interface CommandViewProps {
    client: client.Client
    name: string
    sequenceConfig: client.SequenceConfig
}

export const CommandView = (props: CommandViewProps) => {
    const [currentStepIndex, setCurrentStepIndex] = React.useState<number>(0);
    const [lastExecutedStepIndex, setLastExecutedIndex] = React.useState<number>(0);
    const [executeCommandRsps, setExecuteCommandRsps] = React.useState<client.ExecuteCommandResponse[]>([...Array(props.sequenceConfig.Steps.length)]);
    const [stepInputValues, setStepInputValues] = React.useState<InputValue[][]>([...Array.from({length: props.sequenceConfig.Steps.length}, (v, i) => [])]);
    const [isLoading, setIsLoading] = React.useState<boolean>(true);
    const [count, setCount] = React.useState<number>(0);
    const [updateCount, setUpdateCount] = React.useState<number>(0);
    const stepsCount = props.sequenceConfig.Steps.length

    const firstInputRefs = useRef([]);
    // @ts-ignore
    firstInputRefs.current = [...Array(stepsCount)].map((ref, index) => {
        return React.createRef()
    })

    const stepHeaderRefs = useRef<any>([]);
    // @ts-ignore
    // stepHeaderRefs.current = [...Array(stepsCount)].map((ref, index) => {
    //     return React.createRef()
    // })

    const addToStepHeaderRefs = (el: React.Ref<any>) => {
        if (el && !stepHeaderRefs.current.includes(el)) {
            stepHeaderRefs.current.push(el);
        }
    }

    let stepsBodyRef = useRef<any>(null);

    const setStepInputValue = ((i: number, name: string, value: string) => {
        let stepInputValuesCopy = [...stepInputValues]
        for (let ii = 0; ii < stepInputValuesCopy[i]?.length; ii++) {
            if (stepInputValuesCopy[i][ii].Name === name) {
                stepInputValuesCopy[i][ii].Value = value
                setStepInputValues(stepInputValuesCopy)

                return
            }
        }

        stepInputValuesCopy[i].push({Name: name, Value: value})
        setStepInputValues(stepInputValuesCopy)
    })

    const refresh = () => {
        setCount(count + 1)
    }

    const handleFocus = (stepIndex: number) => {
        setTimeout(() => {
            let blurIndex = stepIndex === 0 ? stepsCount - 1 : stepIndex - 1
            // @ts-ignore
            if (firstInputRefs.current[blurIndex].current != null) {
                // @ts-ignore
                firstInputRefs.current[blurIndex].current.blur()
            }

            // @ts-ignore
            if (firstInputRefs.current[stepIndex].current != null) {
                // @ts-ignore
                firstInputRefs.current[stepIndex].current.focus()
            }
        }, 100)
    }

    useEffect(() => {
        let inputConfigs: client.CommandInputConfig[] | undefined = undefined
        let inputValues: client.InputValue[] | undefined = undefined

        inputConfigs = props.sequenceConfig.Steps[currentStepIndex].Command.Inputs
        inputValues = stepInputValues[currentStepIndex]

        if (inputConfigs && inputConfigs.length !== 0 && inputValues?.length === 0) {
            setIsLoading(false);

            return
        }

        let currentStepIndexCopy = currentStepIndex
        // if (isViewExec) {
        //     currentExecIndexCopy++
        //     setCurrentStepIndex(currentExecIndexCopy)
        //     setExecuteCommandRsps([])
        //     setStepInputValues([...Array.from({length: stepsCount}, (v, i) => [])]);
        //
        //     handleFocus(currentExecIndexCopy)
        //
        //     return
        // }

        setIsLoading(true);

        (async function () {
            let req: client.ExecuteCommandRequest = {
                Slug: props.sequenceConfig.Steps[currentStepIndex].Command.Slug,
                Inputs: inputValues,
            }
            const viewOutputRsp = await props.client.ExecuteCommand(req)

            let executeCommandRspsCopy = [...executeCommandRsps]
            executeCommandRspsCopy[currentStepIndex] = viewOutputRsp

            if (viewOutputRsp.Output.ExitCode == 0) {
                if (currentStepIndexCopy + 1 >= stepsCount) {
                    currentStepIndexCopy = 0
                    // setViewEnv([]);
                } else {
                    currentStepIndexCopy++
                }
            }

            setLastExecutedIndex(currentStepIndex);
            setCurrentStepIndex(currentStepIndexCopy);
            setExecuteCommandRsps(executeCommandRspsCopy);
            setIsLoading(false);
            setUpdateCount(updateCount+1)

            console.log(stepHeaderRefs.current)
            // @ts-ignore
            // stepHeaderRefs.current[currentStepIndex+1]?.scrollIntoView({ behavior: 'smooth', inline: 'center' })

            if (currentStepIndex > 1) {
                stepsBodyRef.current?.scrollTo({
                    // @ts-ignore
                    top: stepHeaderRefs.current[currentStepIndexCopy].offsetTop - 300,
                    behavior: 'smooth'
                });
            }

            handleFocus(currentStepIndexCopy)

            if (viewOutputRsp.Output.ExitCode == 0 && currentStepIndexCopy && (!props.sequenceConfig.Steps[currentStepIndexCopy].Command.Inputs || props.sequenceConfig.Steps[currentStepIndexCopy].Command.Inputs.length === 0)) {
                setTimeout(() => {
                    refresh()
                }, 500)
            }
        })().catch((reason: any) => {
            message.error('failed to get step output: ' + reason);
            setIsLoading(false);
        });
    }, [count]);

    const filename = `${props.name.replaceAll(" ", "_")}_${(new Date()).toISOString().slice(0, 19).replace("T", "_")}.txt`

    return (
        <div className="views__view" key={props.name}>
            <div className="views__view__header">
                <span className="views__view__header__name">
                    {props.name}
                </span>
            </div>
            <div className="views__view__body" ref={stepsBodyRef}>
                <div className="views__view__steps">
                    {props.sequenceConfig.Steps.map((step: client.StepConfig, stepIndex: number) => {
                        let className = "views__view__step"
                        if (stepIndex === currentStepIndex && !isLoading) {
                            className = className + " views__view__step--active"
                        }

                        let rawReq: client.ExecuteCommandRequest = {
                            Slug: props.sequenceConfig.Steps[stepIndex].Command.Slug,
                            Inputs: stepInputValues[stepIndex],
                        }
                        rawReq.Format = client.FormatRaw;

                        let outputClassName = "views__view__output"
                        if (stepIndex == lastExecutedStepIndex && executeCommandRsps[stepIndex] && executeCommandRsps[stepIndex].Output) {
                            outputClassName = outputClassName + " views__view__output--highlight" + updateCount % 2
                        }

                        return (
                            // @ts-ignore
                            <div className={className} key={props.name + stepIndex} ref={addToStepHeaderRefs}>
                                <div className="views__view__step__header">
                                <span className="views__view__step__header__name">
                                    #{stepIndex + 1} {step.Name}
                                </span>
                                    {isLoading && stepIndex === currentStepIndex ?
                                        <span className="views__view__header__loader"><div
                                            className="loader01"/></span> : null}
                                    <span className="views__view__header__raw">
                                    <a rel="noreferrer" onClick={refresh}>Run </a>
                                    <a href={props.client.ExecuteCommandLink(rawReq)} target="_blank"
                                       rel="noreferrer">Raw </a>
                                    <a href={props.client.ExecuteCommandLink(rawReq)} target="_blank" rel="noreferrer"
                                       download={filename}>Download </a>
                                </span>
                                    {step.Command.Inputs ?
                                        <ViewEnv inputConfigs={step.Command.Inputs}
                                                 inputValues={stepInputValues[stepIndex]}
                                                 ref={firstInputRefs.current[stepIndex]}
                                                 setInputValue={(k, v: string) => setStepInputValue(stepIndex, k, v)}
                                                 disabled={stepIndex !== currentStepIndex || isLoading}
                                            // disabled={false}
                                                 refresh={refresh}/> : null}
                                </div>
                                <div
                                    className={outputClassName}>{executeCommandRsps[stepIndex]?.Output?.Stdout ? executeCommandRsps[stepIndex]?.Output?.Stdout : executeCommandRsps[stepIndex]?.Output?.Stderr}</div>
                            </div>
                        )
                    })}
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
