import React from 'react';
import ReactDOM from 'react-dom';
import './index.css';
import App from './App';
// @ts-ignore
import {useRoutes} from 'hookrouter';

const routes = {
    '/': () => <App />,
};

const Root = () => {
    const routeResult = useRoutes(routes);

    return routeResult || 'Page Not Found';
};

ReactDOM.render(<Root/>, document.getElementById('root'));
