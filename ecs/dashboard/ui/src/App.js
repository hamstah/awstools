import React, { Component } from 'react';
import _ from 'lodash'

const API = '/api/overview';

const FontAwesome = ({color, icon, text}) => (
  <div style={{color: color}}>
  <i title={text} className={"fas fa-"+icon}></i>
  </div>
)

const Check = (props) => (
  <FontAwesome icon="check-circle" {...props} />
)

const Pause = (props) => (
  <FontAwesome  icon="pause-circle" {...props}  />
)

class App extends Component {

  constructor(props) {
    super(props)
    this.state = {
      data: {"state": [], "accounts": []},
      refreshInterval: 50 * 1000,
      isLoading: false,
      services: {},
      accounts: [],
      messages: [],
    }
    this.interval = null

}
  tick = () => {
    this.setState({ isLoading: true });

      fetch(API)
        .then(response => {
          if (response.ok) {
            return response.json();
          } else {
            throw new Error('Something went wrong ...');
          }
        })
        .then(data => this.processData(data))
        .catch(error => this.setState({ error, isLoading: false }));
  }

  processData = (data) => {
    var services = {}
    for (const accountState of data["state"]) {
      const prefix = accountState["account"]["prefix"]
        for( const service of accountState["services"]) {
        const name = service["service_name"]
        var clusterName = _.last(service["cluster_arn"].split("/"))
        if (prefix && clusterName.startsWith(prefix )) {
          clusterName = clusterName.slice(prefix.length)
        }
        if (! services[clusterName]) {
          services[clusterName] = {}
        }
        if (! services[clusterName][name]) {
          services[clusterName][name] = {}
        }
        services[clusterName][name][accountState["account"]["account_name"]] = service;

      }
    }
    this.setState({ data, isLoading: false, services, messages: data["messages"] })
  }


  componentDidMount = () => {
   this.interval = setInterval(this.tick, this.state.refreshInterval);
   this.tick()
 }

 componentWillUnmount = () => {
   clearInterval(this.interval);
 }

render = () => {
    const {data, services, messages} = this.state
    const accounts = data["accounts"]
    const subheaders= ["count", "event", "version"];

    var lastSeen = {}

    const vBorderStyle = (index, a) => {
      var l = index == 0 ? "strong-left" : ""
      var r = index == a.length-1 ? "strong-right": ""
      return `${l} ${r}`
    }

    const hBorderStyle = (index, a) => {
      var t = index == 0 ? "strong-top" : ""
      var b = index == a.length-1 ? "strong-bottom": ""
      return `${t} ${b}`
    }

    return (
      <div>
        {messages.length > 0 ? (
      <section className="section">

          <div className="container">
            {messages.map((message, index) => (
              <div className={["notification", "is-small", message["type"] == "error" ? "is-danger" : "is-primary"].join(" ")} key={"message-" + index}>
                {message.message}
              </div>
            ))}
          </div>
        </section>
        ): null}
        <section className="section">

        <div className="container">
          <table className="table is-bordered is-fullwidth is-striped">
            <thead>
              <tr>
                <th className="no-border">
                </th>
                {accounts.map(account => (
                  <th
                    className="strong-left strong-right strong-top"
                    colSpan={subheaders.length}
                    key={"head-"+account["account_name"]}>
                    {account["account_name"]}
                  </th>
                ))}
              </tr>
              <tr>
                <th className="no-border">
                </th>
                {accounts.map(account => (
                  subheaders.map((subhead, index) => (
                    <th
                      className={[vBorderStyle(index, subheaders), "strong-top", "strong-bottom", "subhead"].join(" ")}
                      key={"head-"+account["account_name"]+"-"+subhead}>
                      {subhead}
                    </th>
                  ))
                ))}
              </tr>
            </thead>

            {Object.keys(services).sort().map(clusterName => (
              <tbody key={"body-"+ clusterName}>
                <tr className="divider">
                  <td colSpan={accounts.length * subheaders.length + 1}>
                    {clusterName}
                  </td>
                </tr>
                {Object.keys(services[clusterName]).sort().map((serviceName, rowIndex) => (
                  <tr key={clusterName+"-"+serviceName+"-row"}>
                    <td>
                      {serviceName}
                    </td>
                    {accounts.map( (account, index) => {
                      const s = services[clusterName][serviceName][account["account_name"]]
                      var component = {}
                      if(s) {
                        if(s["status"] == "ACTIVE") {
                          if( s["desired_count"] == s["running_count"]) {
                            component["count"] =
                            <Check color="green" />
                          } else {
                            component["count"] =
                            <FontAwesome
                              color="red"
                              icon="exclamation-triangle"
                              text={s["running_count"] + "/" + s["desired_count"]}/>
                          }
                        } else {
                          component["count"] =
                          <FontAwesome
                            color="red"
                            icon="pause-circle"
                            text={s["status"]}/>
                        }

                        var ttd = {}
                        for(var accountState of data["state"]) {
                          if (accountState["account"]["account_name"] === account["account_name"]) {
                            ttd = accountState["task_definitions"]
                          }
                        }
                        const td = ttd[s["task_definition"]]
                        if ( td ) {
                          var images = []
                          var allOk = true
                          for (var container of td["container_definitions"]) {

                            const key = clusterName + "|" + s["service_name"] + "|" + container["name"]
                            var ok = true
                            var tag = _.last(container["image"].split("/"))
                            if( lastSeen[key] && lastSeen[key] !== tag) {
                              ok = false
                              allOk = false
                            }
                            lastSeen[key] = tag
                            var tag = _.last(container["image"].split("/")).split(":")
                            var lastTag = _.last(tag)
                            if (lastTag.length > 7) {
                              lastTag = lastTag.slice(0, 7) + ".."
                              tag[tag.length-1] = lastTag
                            }
                            tag = tag.join(":")
                            images.push([tag, container["image"], ok])
                          }
                          component["version"] = ( allOk ?
                            <Check color="green" text={images.map(image => image[0]).join("\n")} />
                            :
                            <FontAwesome
                              color="orange"
                              icon="exclamation-triangle"
                              text={images.map(image => image[0]).join("\n")} />
                          )
                        } else {
                          component["version"] =
                          <FontAwesome
                            color="red"
                            icon="exclamation-triangle"
                            text="Can't find task definition" />
                        }

                        const event =  s["events"][0]
                        if ( event["message"].indexOf("steady state") != -1) {
                          component["event"] =
                          <Check color="green" />
                        } else {
                          component["event"] =
                          <FontAwesome
                            color="red"
                            icon="exclamation-triangle"
                            text={event["created_at"] + ": " + event["message"]} />
                        }
                      }

                      return ( subheaders.map((subhead, colIndex) => (
                        <td
                          className={[vBorderStyle(colIndex, subheaders), hBorderStyle(rowIndex, Object.keys(services[clusterName]))].join(" ")}
                          style={{textAlign: "center", backgroundColor: component[subhead] ? "transparent" : "#eee"}}
                          key={clusterName+"-"+serviceName+"-cell-"+account["account_name"]+"-"+subhead}>
                          {component[subhead]}
                        </td>
                      ))
                    )})}
                  </tr>
                ))}
              </tbody>
            )
          )}

        </table>
      </div>
    </section>
  </div>
    )
  }
}


export default App;
