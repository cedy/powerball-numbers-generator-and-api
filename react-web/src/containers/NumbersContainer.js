import React from 'react';
import Numbers from '../components/Numbers';
import { w3cwebsocket as W3CWebSocket } from "websocket";
import PPagination from '../components/Pagination';

const client = new W3CWebSocket('ws://127.0.0.1:8080/ws');
const API_ENDPOINT = "http://127.0.0.1:8080/top/count/";


class NumbersContainer extends React.Component {
    constructor(props) {
        super(props);
        this.onClickFirstPage = this.onClickFirstPage.bind(this);
        this.onClickNextPage = this.onClickNextPage.bind(this);
        this.onClickPreviousPage = this.onClickPreviousPage.bind(this);
        this.state = {
            numbersList: {},
            currentPage: 1,
            
        };
    }
        
    componentDidMount() {
        this.loadPage(this.state.currentPage);
    }

    loadPage(page) {
        fetch(API_ENDPOINT + page)
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

    onClickFirstPage(event) {
        if (this.state.currentPage === 1) {
            return
        }
        this.setState({currentPage: 1});
        this.loadPage(1);
    }

    onClickNextPage(event) {
        if (Object.keys(this.state.numbersList).length < 100) {
            return
        }
        let newPage = this.state.currentPage + 1
        this.setState({currentPage: newPage});
        this.loadPage(newPage);
    }

    onClickPreviousPage(event) {
        if (this.state.currentPage === 1) {
            return
        }
        let currentPage = this.state.currentPage - 1 ;
        this.setState({currentPage: currentPage});
        this.loadPage(currentPage);
    }

    render () {
        return (
            <div className="main-container">
                {
                    Object.keys(this.state.numbersList).map((numbers) => (
                        <Numbers key={numbers} numbers={numbers} count={this.state.numbersList[numbers]} animate={true} />
                    ))
                }
            
            <div className="d-flex mt-4 justify-content-center">
                <PPagination
                    currentPage={this.state.currentPage}
                    first={this.onClickFirstPage}
                    previous={this.onClickPreviousPage}
                    next={this.onClickNextPage}
                />
            </div>
            </div>
        )
    }
}

export default NumbersContainer;
