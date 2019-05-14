import React from 'react'
import './App.css';
import { withStyles } from '@material-ui/core/styles';
import Grid from '@material-ui/core/Grid';
import Typography from '@material-ui/core/Typography';
import AttachFileIcon from '@material-ui/icons/AttachFile';
import Dropzone from 'react-dropzone'
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import Paper from '@material-ui/core/Paper';
import Highlight from 'react-highlight'
import ExpansionPanel from '@material-ui/core/ExpansionPanel';
import ExpansionPanelDetails from '@material-ui/core/ExpansionPanelDetails';
import ExpansionPanelSummary from '@material-ui/core/ExpansionPanelSummary';
import ExpandMoreIcon from '@material-ui/icons/ExpandMore';
import { SnackbarProvider, withSnackbar } from 'notistack';
import SpeedDial from '@material-ui/lab/SpeedDial';
import SpeedDialIcon from '@material-ui/lab/SpeedDialIcon';
import SpeedDialAction from '@material-ui/lab/SpeedDialAction';
import { unstable_Box as Box } from '@material-ui/core/Box';

import VTable from './table'
import Log from './log'


const styles = theme => ({
  root: {
    flexGrow: 1,
  },
  grow: {
    flexGrow: 1,
  },
  menuButton: {
    marginLeft: -12,
    marginRight: 20,
  },
  list: {
    width: '100%',
    overflowX: 'auto',
  },
  speedDial: {
    zIndex: 100000,
    position: 'fixed',
    bottom: theme.spacing.unit * 2,
    right: theme.spacing.unit * 3,
  },
});


class RecordItem extends React.Component {
  state = {
    opened: false,
  }

  onChangeExpand(e, expanded) {
    this.setState({opened: expanded})
  }

  render() {
    const { classes } = this.props;
    const { record } = this.props;

    return (
      <TableCell className={classes.listTableTd}>
          <ExpansionPanel className={classes.recorditem} onChange={(e, expaned) => this.onChangeExpand(e, expaned)}>
            <ExpansionPanelSummary
                expandIcon={<ExpandMoreIcon className={classes.recorditemIcon} />}
                className={classes.recorditemTitle}>
              <div className={classes.column}>
                <Typography className={classes.heading}>{record.message}</Typography>
              </div>
            </ExpansionPanelSummary>
            <ExpansionPanelDetails className={classes.details + ' ' + classes.recorditemDetails}>
              {this.state.opened ? (
                <Highlight className='json'>{JSON.stringify(record.extra, null, 2)}</Highlight>
              ) : (
                <span />
              )}
            </ExpansionPanelDetails>
          </ExpansionPanel>
      </TableCell>
    )
  }
}

class CenteredGrid extends React.Component {
  state = {
    menu: false,
    fileDialog: false,
    bottom: true,
    records: [],
    nodes: [],
    record: null,
    speedDial: false,
  }

  toggleDrawer = (open) => () => {
    this.setState({
      'menu': open,
    });
  }

  toggleBottom = (open) => () => {
    this.setState({
      'bottom': open,
    });
  }

  onSelectedFile = (acceptedFiles) => {
    const reader = new FileReader();

    reader.onabort = () => console.log('file reading was aborted');
    reader.onerror = () => console.log('file reading has failed');
    reader.onload = () => {
      const binaryStr = reader.result

      try {
        var log = Log.load(binaryStr)
      } catch(e) {
        this.props.enqueueSnackbar('failed to load log file,' + acceptedFiles, {variant: 'error'})
        return
      }

      this.setState({records: log.records})
      this.setState({nodes: log.nodes})

      this.props.enqueueSnackbar(
        'log file,' + acceptedFiles + ' successfully imported: ' + log.records.length + ' records found',
        {variant: 'info'},
      )
    }

    acceptedFiles.forEach(file => reader.readAsBinaryString(file))
  }


  handleSpeedDialOpen = () => {
    this.setState({ speedDial: true, });
  };

  handleSpeedDialClose = () => {
    this.setState({ speedDial: false, });
  };

  componentDidUpdate() {
    setTimeout(e => {
      var tds = null
      try {
        tds = document.getElementsByTagName('table')[1].getElementsByTagName('tbody')[0].getElementsByTagName('tr')[0].getElementsByTagName('td')
      } catch {
        return
      }


      var fixed_ths = document.getElementsByTagName('table')[0].getElementsByTagName('thead')[0].getElementsByTagName('tr')[0].getElementsByTagName('th')

      Array.from(tds).map((e, i) => {
        fixed_ths[i].style.width = e.offsetWidth + 'px'
        e.style.width = fixed_ths[i].style.width
      })
    }, 1000)
  }

  render() {
    const { classes } = this.props;

    return <div className={classes.root}>
      <div className={classes.root}>

        <div style={{display: 'none'}}>
          <Dropzone ref='dropzone' onDrop={acceptedFiles => this.onSelectedFile(acceptedFiles)}>
            {({getRootProps, getInputProps}) => (
              <section>
                <div {...getRootProps()}>
                  <input {...getInputProps()} />
                  <div>Drag 'n' drop some files here, or click to select files</div>
                </div>
              </section>
            )}
          </Dropzone>
        </div>

      </div>

        <Table className={' fixed'}>
          <TableHead>
            <TableRow>
              <TableCell className={classes.listTableT} key={'t'}><div>T</div></TableCell>
              {this.state.nodes.map(node => (
                <TableCell align="left" key={node}><div>{node}</div></TableCell>
              ))}
            </TableRow>
          </TableHead>
        </Table>

      <Box height='100%'>
        <Table className={' scrollable'}>
          <TableBody>
            {this.state.records.map(record => {
              var i = this.state.nodes.indexOf(record.node)
              if (i < 0) {
                return null
              }

              return (
                <TableRow key={record.id}>
                  <TableCell key={record.id + '-t'}>
                    {record.t.elapsed(this.state.records[0].t)}
                  </TableCell>
                  {this.state.nodes.map((node, index) => (
                    i === index ? (
                      <RecordItem classes={classes} record={record} key={record.id + '-' + record.node} />
                    ) : (
                      <TableCell key={record.id + '-' + node}></TableCell>
                    )
                  ))}
                </TableRow>
              )
            })}
          </TableBody>
        </Table>
      </Box>

        <SpeedDial
          ariaLabel="SpeedDial tooltip example"
          className={classes.speedDial}
          icon={<SpeedDialIcon />}
          onBlur={this.handleSpeedDialClose}
          onClick={this.handleSpeedDialClick}
          onClose={this.handleSpeedDialClose}
          onFocus={this.handleSpeedDialOpen}
          onMouseEnter={this.handleSpeedDialOpen}
          onMouseLeave={this.handleSpeedDialClose}
          open={this.state.speedDial}
        >
            <SpeedDialAction
              key={'import-log-file'}
              icon={<AttachFileIcon />}
              tooltipTitle={'import new log'}
              tooltipOpen
              onClick={e=>{this.refs['dropzone'].open()}}
            />
        </SpeedDial>


    </div>
  }
}

const MyApp = withSnackbar(CenteredGrid);

class IntegrationNotistack extends React.Component {
  render() {
    return (
      <SnackbarProvider maxSnack={3} autoHideDuration={2000}>
        <MyApp {...this.props} />
      </SnackbarProvider>
    )
  }
}

export default withStyles(styles)(IntegrationNotistack)
