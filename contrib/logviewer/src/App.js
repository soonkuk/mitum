import React from 'react'
import ReactDOM from 'react-dom';
import './App.css';
import { withStyles } from '@material-ui/core/styles';
import Typography from '@material-ui/core/Typography';
import AttachFileIcon from '@material-ui/icons/AttachFile';
import Dropzone from 'react-dropzone'
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableHead from '@material-ui/core/TableHead';
import Grid from '@material-ui/core/Grid';
import TableRow from '@material-ui/core/TableRow';
import Highlight from 'react-highlight'
import { SnackbarProvider, withSnackbar } from 'notistack';
import SpeedDial from '@material-ui/lab/SpeedDial';
import SpeedDialIcon from '@material-ui/lab/SpeedDialIcon';
import IconButton from '@material-ui/core/IconButton';
import SettingsOverscanIcon from '@material-ui/icons/SettingsOverscan';
import ChildCareIcon from '@material-ui/icons/ChildCare';
import BookmarksIcon from '@material-ui/icons/Bookmarks';
import SpeedDialAction from '@material-ui/lab/SpeedDialAction';
import { unstable_Box as Box } from '@material-ui/core/Box';
import Chip from '@material-ui/core/Chip';

import Log from './log'
import raw from './raw'


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
    var promises = []
    for (let file of acceptedFiles) {
      var p = new Promise(function(resolve, reject) {
        var reader = new FileReader();
        reader.onload = () => {
          resolve(reader.result)
        }

        reader.readAsBinaryString(file)
      })
      promises.push(p)
    }

    Promise.all(promises).then(values => {
      var result = ''.concat(...values)

      try {
        var log = Log.load(result)
      } catch(e) {
        this.props.enqueueSnackbar('failed to load logs', {variant: 'error'})
        return
      }

      this.setState({records: log.records})
      this.setState({nodes: log.nodes})

      this.props.enqueueSnackbar(
        'logs successfully imported: ' + log.records.length + ' records found',
        {variant: 'info'},
      )

    })
  }


  handleSpeedDialOpen = () => {
    this.setState({ speedDial: true, });
  };

  handleSpeedDialClose = () => {
    this.setState({ speedDial: false, });
  };

  toggleExpandAll = () => {
    const node = ReactDOM.findDOMNode(this)
    if (! node instanceof HTMLElement) {
      return
    }

    const children = node.querySelectorAll('.row-detail')
    Array.from(children).map(c => {
      c.classList.toggle('row-detail-open')
      return null
    })
  }

  importTestData = () => {
      var log = Log.load(raw)
      this.setState({records: log.records})
      this.setState({nodes: log.nodes})

      this.props.enqueueSnackbar(
        'test log data successfully imported: ' + log.records.length + ' records found',
        {variant: 'info'},
      )
  }

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
        if (fixed_ths[i] === undefined) {
          return null
        }

        fixed_ths[i].style.width = e.offsetWidth + 'px'
        e.style.width = fixed_ths[i].style.width
        return null
      })

    }, 1000)
  }

  openDetail(ref, open) {
    const tr = ReactDOM.findDOMNode(ref.current)
    if (open === undefined) {
      tr.nextSibling.classList.toggle('row-detail-open')
      return
    }

    if (open === true) {
      tr.nextSibling.classList.add('row-detail-open')
    } else {
      tr.nextSibling.classList.remove('row-detail-open')
    }
    return
  }

  shouldComponentUpdate(nextProps, nextState) {
    return true
  }

  renderRecord(first, record, nodes) {
    const { classes } = this.props;

    var i = nodes.indexOf(record.node)
    if (i < 0) {
      return null
    }

    var rowRef = React.createRef()
    var rowDetailRef = React.createRef()

    return <React.Fragment key={record.id + 'f'}>
      <TableRow key={record.id} ref={rowRef}>
        <TableCell key={record.id + '-t'}>
          <IconButton className={classes.button} aria-label='Bookmark' onClick={e => {
            const tr = ReactDOM.findDOMNode(rowRef.current)
            tr.classList.toggle('selected')
            tr.nextSibling.classList.toggle('selected')

            this.openDetail(rowRef, true)
          }}>
            <BookmarksIcon />
          </IconButton>
          <Chip label={record.level} className={'lvl lvl-' + record.level} color='secondary' />
          <span className='t'>
            {record.t.elapsed(first.t)}
          </span>
        </TableCell>
        {nodes.map((node, index) => (
          <TableCell
            className={classes.listTableTd}
            key={record.id + node + '-m'}
            onClick={e => this.openDetail(rowRef)}
          >
          {i === index ? (
              <Typography key={record.id + node + 'ty'}>{record.message}</Typography>
            ) : (
              <Typography key={record.id + node + 'ty'}></Typography>
            )
          }
          </TableCell>
        ))}
      </TableRow>
      <RecordDetail classes={classes} nodes={nodes} record={record} ref={rowDetailRef} />
    </React.Fragment>
  }

  renderRecords(records, nodes) {
    const { classes } = this.props;

    if (this.state.records.length < 1) {
      return <React.Fragment>
        <Table className={' fixed'}>
          <TableHead>
            <TableRow>
              <TableCell className={classes.listTableT} key={'t'}><div>T</div></TableCell>
                <TableCell key={'tc-0'}></TableCell>
                <TableCell key={'tc-1'}></TableCell>
                <TableCell key={'tc-2'}></TableCell>
                <TableCell key={'tc-3'}></TableCell>
            </TableRow>
          </TableHead>
        </Table>
      </React.Fragment>
    }

    return <React.Fragment>
      <Table className={' fixed'}>
        <TableHead>
          <TableRow>
            <TableCell className={classes.listTableT} key={'t'}><div>T</div></TableCell>
            {this.state.nodes.map(node => (
              <TableCell align='left' key={node}><div>{node}</div></TableCell>
            ))}
          </TableRow>
        </TableHead>
      </Table>

      <Box height='100%'>
        <Table className={' scrollable'}>
          <TableBody>
            <TableRow>
                <TableCell key={'h'}>.</TableCell>
              {this.state.nodes.map((node, index) => (
                <TableCell key={node + 'h'}></TableCell>
              ))}
            </TableRow>
            {this.state.records.map(record => {
              return this.renderRecord(records[0], record, nodes)
            })}
          </TableBody>
        </Table>
      </Box>
    </React.Fragment>
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

      {this.renderRecords(this.state.records, this.state.nodes)}
      <SpeedDial
        ariaLabel='SpeedDial tooltip example'
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
        <SpeedDialAction
          key={'expand-collapse-all'}
          icon={<SettingsOverscanIcon />}
          tooltipTitle={'expand/collapse all'}
          tooltipOpen
          onClick={e=>{this.toggleExpandAll()}}
        />
        <SpeedDialAction
          key={'test data'}
          icon={<ChildCareIcon />}
          tooltipTitle={'test data'}
          tooltipOpen
          onClick={e=>{this.importTestData()}}
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

class RecordDetail extends React.Component {
  render() {
    const { record, nodes, classes } = this.props;

    return <TableRow className={'row-detail'} key={record.id + 'detail'}>
      <TableCell colSpan={nodes.length + 1}>
        <Grid container className={classes.root} spacing={16}>
          <Grid item xs={4}>
            <Highlight className='json'>{JSON.stringify(record.basic(), null, 2)}</Highlight>
          </Grid>
          <Grid item xs={8}>
            <Highlight className='json'>{JSON.stringify(record.extra, null, 2)}</Highlight>
          </Grid>
        </Grid>
      </TableCell>
    </TableRow>
  }
}

export default withStyles(styles)(IntegrationNotistack)
