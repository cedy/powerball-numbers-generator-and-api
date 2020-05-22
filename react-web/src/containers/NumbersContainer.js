import React from 'react';
import Numbers from '../components/Numbers';
import { w3cwebsocket as W3CWebSocket } from "websocket";

const client = new W3CWebSocket('ws://localhost:8080/ws');
const API_ENDPOINT = "http://localhost:8080/top/count/0";


class NumbersContainer extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            numbersList: {},
        };
    }
        
    componentDidMount() {
        fetch(API_ENDPOINT)
        .then((resonse) => resonse.json())
            .then((response_json) => {
                let numbersList = {};
                response_json.forEach(struct => {
                    numbersList[struct.numbers] = struct.count;
                });
                this.setState({numbersList: numbersList});
                return true
            })
            .then((ready) => {
                if (ready) {
                client.onopen = () => {
                    console.log("WebSocket Client Connected");
                };
                client.onmessage = (evt) => {
                    this.addNumbers(evt.data);
                };
                }
            })
            .catch((error) => alert("Something went wrong, please refresh the page"));
    }

    componentWillUnmount() {
        client.onclose = (evt) => {
            console.log("Connection closed");
        };
    }

    addNumbers(numbers) {
        // limit 100 records on the page 
        if ((Object.keys(this.state.numbersList).length >= 100) && !(numbers in this.state.numbersList)) {
            return
        }
        let numbersList = Object.assign({}, this.state.numbersList);
        if (numbers in this.state.numbersList){
            numbersList[numbers]++;
        } else {
            numbersList[numbers] = 1;
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
