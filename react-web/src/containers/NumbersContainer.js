import React from 'react';
import Numbers from '../components/Numbers';
import { w3cwebsocket as W3CWebSocket } from "websocket";

const client = new W3CWebSocket('ws://localhost:8080/ws');

class NumbersContainer extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            numbersList: {},
        };
    }
        
    componentDidMount() {
        client.onopen = () => {
            console.log("WebSocket Client Connected");
        };
        client.onmessage = (evt) => {
            this.addNumbers(evt.data);
        };
    }

    componentWillUnmount() {
        client.onclose = (evt) => {
            console.log("Connection closed");
        };
    }

    addNumbers(message) {
        // limit 100 records on the page 
        if ((Object.keys(this.state.numbersList).length > 100) && !(message.slice(0,7) in this.state.numbersList)) {
            return
        }
        let numbersList = Object.assign({}, this.state.numbersList);
        var first = message.slice(0,7);
        if (first in this.state.numbersList){
            numbersList[first]++;
        } else {
            numbersList[first] = 1;
        }
        this.setState({numbersList: numbersList});
    }

    render () {
        return (
            <div className="main-container">
                {
                    Object.keys(this.state.numbersList).map((numbers) => (
                        <Numbers key={numbers} numbers={numbers} count={this.state.numbersList[numbers]} animate={true} />
                    ))
                }
            </div>
        )
    }
}

export default NumbersContainer;
