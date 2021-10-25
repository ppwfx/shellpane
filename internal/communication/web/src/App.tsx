import React, {useEffect} from 'react';
import './App.css';

import * as client from './client';
import {Col, message, Row} from 'antd';
import SequenceView from './components/SequenceView'

let c = new client.Client({
    config: {
        addr: (process.env.REACT_APP_SHELLPANE_HOST ? process.env.REACT_APP_SHELLPANE_HOST : window.location.origin),
    },
});

function App() {
    const [getViewConfigsRsp, setViewConfigsRsp] = React.useState<client.GetViewConfigsResponse | undefined>(undefined);
    const [getCategoryConfigsRsp, setCategoryConfigsRsp] = React.useState<client.GetCategoryConfigsResponse | undefined>(undefined);
    const [category, setCategory] = React.useState<client.CategoryConfig | undefined>(undefined);

    useEffect(() => {
        (async function () {
            const rsp = await c.GetViewConfigs({})

            setViewConfigsRsp(rsp);
            console.log(rsp)
        })().catch((reason: any) => {
            message.error('failed to get view configs: ' + reason.message, 5);
            setViewConfigsRsp(undefined);
        });

        (async function () {
            const rsp = await c.GetCategoryConfigs({})

            setCategoryConfigsRsp(rsp);
            setCategory(rsp && rsp.CategoryConfigs && rsp.CategoryConfigs[0] ? rsp.CategoryConfigs[0]: undefined)
            console.log(rsp)
        })().catch((reason: any) => {
            message.error('failed to get category configs: ' + reason.message, 5);
            setCategoryConfigsRsp(undefined);
        });
    }, []);

    let viewConfigs:client.ViewConfig[]
    if (category) {
        viewConfigs = getViewConfigsRsp?.ViewConfigs?.filter((c) => c.Category.Slug == category.Slug) || []
    } else {
        viewConfigs = getViewConfigsRsp?.ViewConfigs || []
    }

    const addAlpha = (color:string, opacity:number):string => {
        // coerce values so ti is between 0 and 1.
        var _opacity = Math.round(Math.min(Math.max(opacity || 1, 0), 1) * 255);
        return color + _opacity.toString(16).toUpperCase();
    }

    return (<>
            <div className="header">
                <span className={"header__logo"}>🐚 shellpane{category ? "/":""}{category ? <span className={"header__logo__category"} style={{color:category.Color}}>{category.Name}</span>:null}</span>
                <span className={"header__categories"}>
                    {getCategoryConfigsRsp?.CategoryConfigs?.map((c => {
                        return <a className={"header__category__category" + (c.Slug == category?.Slug ? " header__categories__category--selected": "")} style={{color:c.Color}} onClick={()=>setCategory(c)}> {c.Name}</a>
                    }))}
                </span>
            </div>
            {!category ? null: <div className={"App"}>
                <Row gutter={[32, 32]}>
                    {viewConfigs.map((v: client.ViewConfig) => {
                        if (v.Sequence) {
                            return <Col key={v.Name} span={12}><SequenceView client={c} name={v.Name} viewConfig={v}
                                                                            sequenceConfig={v.Sequence}/></Col>
                        }
                    })}
                </Row>
            </div>}

        </>
    );
}

export default App;
