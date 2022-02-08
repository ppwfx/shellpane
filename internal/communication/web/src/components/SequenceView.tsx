import React, {useEffect, useRef} from 'react';
import * as client from '../client';
import {message} from 'antd';
import ReactECharts from 'echarts-for-react';
import Chart from "react-apexcharts";

interface SequenceViewProps {
    client: client.Client
    name: string
    viewConfig: client.ViewConfig
    sequenceConfig: client.SequenceConfig
}

export const SequenceView = (props: SequenceViewProps) => {
    const [currentStepIndex, setCurrentStepIndex] = React.useState<number>(0);
    const [lastExecutedStepIndex, setLastExecutedIndex] = React.useState<number>(0);
    const [executeCommandRsps, setExecuteCommandRsps] = React.useState<client.ExecuteCommandResponse[]>([...Array(props.sequenceConfig.Steps.length)]);
    const [values, setValues] = React.useState<{ [name: string]: string }>({});
    const [isLoading, setIsLoading] = React.useState<boolean>(true);
    const [count, setCount] = React.useState<number>(0);
    const [updateCount, setUpdateCount] = React.useState<number>(0);
    const stepsCount = props.sequenceConfig.Steps.length

    let seenInputs: { [name: string]: boolean } = {}
    const inputConfigs: client.CommandInputConfig[][] = props.sequenceConfig.Steps.map((s: client.StepConfig, i: number) => {
        let stepInputs: client.CommandInputConfig[] = []

        if (!s.Command.Inputs) {
            return []
        }

        s.Command.Inputs.forEach(((input) => {
            if (seenInputs[input.Input.Slug]) {
                return
            }

            seenInputs[input.Input.Slug] = true
            stepInputs.push(input)
        }))

        return stepInputs
    })

    const firstInputRefs = useRef([]);
    // @ts-ignore
    firstInputRefs.current = [...Array(stepsCount)].map((ref, index) => {
        return React.createRef()
    })

    const stepHeaderRefs = useRef<any>([]);
    const addToStepHeaderRefs = (el: React.Ref<any>) => {
        if (el && !stepHeaderRefs.current.includes(el)) {
            stepHeaderRefs.current.push(el);
        }
    }

    let stepsBodyRef = useRef<any>(null);

    const setStepInputValue = ((i: number, name: string, value: string) => {
        let valuesCopy = Object.assign({}, values);
        valuesCopy[name] = value
        setValues(valuesCopy)
    })

    const refresh = () => {
        setCount(count + 1)
    }

    const toInputValues = (
        commandInputsConfigs: client.CommandInputConfig[],
        values: { [name: string]: string }
    ): client.InputValue[] => {
        if (!commandInputsConfigs) {
            return []
        }

        let inputValues: client.InputValue[] = []
        commandInputsConfigs.forEach((c) => {
            if (!values[c.Input.Slug]) {
                return
            }

            inputValues.push({
                Name: c.Input.Slug,
                Value: values[c.Input.Slug],
            })
        })

        return inputValues
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
        }, 0)
    }

    useEffect(() => {
        let inputValues = toInputValues(props.sequenceConfig.Steps[currentStepIndex].Command.Inputs, values)

        let executeCommandRspsCopy = [...executeCommandRsps]
        if (currentStepIndex == 0) {
            setExecuteCommandRsps([...Array(props.sequenceConfig.Steps.length)]);
            executeCommandRspsCopy = [...Array(props.sequenceConfig.Steps.length)]

            let firstStepValuesOnly: { [name: string]: string } = {}
            inputConfigs[0].forEach((commandInputConfig => {
                firstStepValuesOnly[commandInputConfig.Input.Slug] = values[commandInputConfig.Input.Slug]
            }))

            setValues(firstStepValuesOnly)
        }

        if (inputConfigs[currentStepIndex] && inputConfigs[currentStepIndex].length !== 0 && inputValues?.length === 0) {
            setIsLoading(false);

            return
        }

        let currentStepIndexCopy = currentStepIndex

        setIsLoading(true);

        (async function () {
            let req: client.ExecuteCommandRequest = {
                Slug: props.sequenceConfig.Steps[currentStepIndex].Command.Slug,
                Inputs: inputValues,
            }
            const viewOutputRsp = await props.client.ExecuteCommand(req)

            executeCommandRspsCopy[currentStepIndex] = viewOutputRsp

            if (viewOutputRsp.Output.ExitCode == 0) {
                if (currentStepIndexCopy + 1 >= stepsCount) {
                    currentStepIndexCopy = 0
                    if (inputConfigs[0].length != 0) {
                        let valuesCopy = Object.assign({}, values)

                        inputConfigs[0].forEach((commandInputConfig => {
                            valuesCopy[commandInputConfig.Input.Slug] = ""
                        }))

                        setValues(valuesCopy)
                    }
                } else {
                    currentStepIndexCopy++
                }
            }

            setLastExecutedIndex(currentStepIndex);
            setCurrentStepIndex(currentStepIndexCopy);
            setExecuteCommandRsps(executeCommandRspsCopy);
            setIsLoading(false);
            setUpdateCount(updateCount + 1)

            if (currentStepIndex > 1) {
                stepsBodyRef.current?.scrollTo({
                    // @ts-ignore
                    top: stepHeaderRefs.current[currentStepIndexCopy].offsetTop - 300,
                    behavior: 'smooth'
                });
            }

            handleFocus(currentStepIndexCopy)

            if (viewOutputRsp.Output.ExitCode == 0 && currentStepIndexCopy && (!inputConfigs[currentStepIndexCopy] || inputConfigs[currentStepIndexCopy].length === 0)) {
                setTimeout(() => {
                    refresh()
                }, 0)
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
            <div className={"views__view__body scrollbar-color--" + props.viewConfig.Category.Slug} ref={stepsBodyRef}>
                <div className="views__view__steps">
                    {props.sequenceConfig.Steps.map((step: client.StepConfig, stepIndex: number) => {
                        let className = "views__view__step"
                        if (stepIndex === currentStepIndex && !isLoading) {
                            className = className + " a-color--" + props.viewConfig.Category.Slug + " input-background--" + props.viewConfig.Category.Slug
                        } else {
                            className = className + " views__view__step--disabled"
                        }

                        let inputValues = toInputValues(props.sequenceConfig.Steps[stepIndex].Command.Inputs, values)
                        let rawReq: client.ExecuteCommandRequest = {
                            Slug: props.sequenceConfig.Steps[stepIndex].Command.Slug,
                            Inputs: inputValues,
                        }
                        rawReq.Format = client.FormatRaw;

                        let outputClassName = "views__view__output"
                        if (stepIndex == lastExecutedStepIndex && executeCommandRsps[stepIndex] && executeCommandRsps[stepIndex].Output) {
                            outputClassName = outputClassName + " flash-border--" + props.viewConfig.Category.Slug + updateCount % 2
                        }

                        let output = null;

                        switch (step.Command.Display) {
                            case "echarts-json":
                                if (!executeCommandRsps[stepIndex]?.Output) {
                                    break
                                }

                                let option = JSON.parse(executeCommandRsps[stepIndex]?.Output?.Stdout)

                                output= <ReactECharts option={option}/>

                                break
                            case "apexcharts-json":
                                if (!executeCommandRsps[stepIndex]?.Output) {
                                    break
                                }

                                let options = JSON.parse(executeCommandRsps[stepIndex]?.Output?.Stdout)

                                output= <Chart options={options} series={options.series} height={options.chart?.height} type={options.chart?.type} width={options.chart?.width}/>

                                break
                            default:
                                output= executeCommandRsps[stepIndex]?.Output?.Stdout ? executeCommandRsps[stepIndex]?.Output?.Stdout : executeCommandRsps[stepIndex]?.Output?.Stderr
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
                                            className={"loader--" + props.viewConfig.Category.Slug}/></span> : null}
                                    <span className="views__view__header__raw">
                                    <a rel="noreferrer" onClick={refresh}>Run </a>
                                    <a href={props.client.ExecuteCommandLink(rawReq)} target="_blank"
                                       rel="noreferrer">Raw </a>
                                    <a href={props.client.ExecuteCommandLink(rawReq)} target="_blank" rel="noreferrer"
                                       download={filename}>Download </a>
                                </span>
                                    {inputConfigs[stepIndex] ?
                                        <ViewEnv inputConfigs={inputConfigs[stepIndex]}
                                                 inputValues={inputValues}
                                                 ref={firstInputRefs.current[stepIndex]}
                                                 setInputValue={(k, v: string) => setStepInputValue(stepIndex, k, v)}
                                                 disabled={stepIndex !== currentStepIndex || isLoading}
                                            // disabled={false}
                                                 refresh={refresh}/> : null}
                                </div>
                                <div className={outputClassName}>
                                    {output}
                                </div>
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


export default SequenceView;
