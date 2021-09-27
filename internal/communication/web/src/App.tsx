import React, {useEffect} from 'react';
import './App.css';
import * as client from './client';
import {Col, Row, message} from 'antd';
import View from './components/View'

let c = new client.Client({
    config: {
        addr: (process.env.REACT_APP_SHELLPANE_HOST ? process.env.REACT_APP_SHELLPANE_HOST : window.location.origin),
    },
});

console.log(process.env)

function App() {
    const [viewSpecsRsp, setViewSpecsRsp] = React.useState<client.GetViewSpecsResponse | undefined>(undefined);

    useEffect(() => {
        (async function () {
            const viewSpecsRsp = await c.GetViewSpecs({})

            setViewSpecsRsp(viewSpecsRsp);
        })().catch((reason: any) => {
            message.error('failed to get view specs: ' + reason.message, 5);
            setViewSpecsRsp(undefined);
        });
    }, []);

    return (
        <div className="App">
            <div className="header">shellpane</div>
            <Row gutter={[16, 16]}>
                {viewSpecsRsp?.Specs?.map((s: client.ViewSpec) => <Col key={s.Name} span={12}><View client={c} viewSpec={s}/></Col>)}
            </Row>
        </div>
    );
}

export default App;
