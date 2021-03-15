# CloudODM

![cloudodm-64x64](https://user-images.githubusercontent.com/1951843/51078515-02348000-1684-11e9-8f96-ed056b0cbe98.png)

A command line tool to process aerial imagery in the cloud via [NodeODM](https://github.com/OpenDroneMap/NodeODM)'s API.

![image](https://user-images.githubusercontent.com/1951843/51078579-3492ad00-1685-11e9-9fcd-0beda36ad56b.png)

## Getting Started

1. [Download the application](https://github.com/OpenDroneMap/CloudODM/releases) for Windows, Mac or Linux.
2. Extract the application in a folder of your choice (for example, `c:\odm`).
3. Open a command prompt and navigate to the folder (open the "Command Prompt" application, then `cd \odm`).
4. Run `odm c:\path\to\images --dsm`.

This command will process all the images in the directory `c:\path\to\images` and save the results (including an orthophoto, a point cloud, a 3D model and a digital surface model) to `c:\odm\output`. You can pass more options for processing by appending them at the end of the command. To see a list of options, simply issue:

`odm args`

See `odm --help` for more options.

## Using GCPs

To include a GCP for additional georeferencing accuracy, simply create a .txt file according to the [Ground Control Points format specification](https://docs.opendronemap.org/gcp.html#gcp-file-format) and place it along with the images.

## Processing Node Management

By default CloudODM will randomly choose a default node from the list of [publicly available nodes](https://github.com/OpenDroneMap/CloudODM/blob/master/public_nodes.json). If you are running your own processing node via [NodeODM](https://github.com/OpenDroneMap/NodeODM) you can add a node by running the following:

`odm node add mynode http://address:port`

Then run odm as following:

`odm -n mynode c:\path\to\images`

If no node is specified, the `default` node is selected. To see a list of nodes you can run:

`odm node -v`

For more information run `odm node --help`.

If you are interested in adding your node to the list of [public nodes](https://github.com/OpenDroneMap/CloudODM/blob/master/public_nodes.json) please open an [issue](https://github.com/OpenDroneMap/CloudODM/issues).

## Running From Sources

```bash
go get -u github.com/OpenDroneMap/CloudODM
go run github.com/OpenDroneMap/CloudODM/cmd/odm --help
```

## Building From Sources

We use [Goreleaser](https://goreleaser.com/) to build and deploy CloudODM. See Goreleaser's [documentation](https://goreleaser.com/) for installation and deployment instructions. You should just need to install the `goreleaser` application and then run:

`goreleaser release --skip-publish --snapshot`

## Reporting Issues / Feature Requests / Feedback

Please open an [issue](https://github.com/OpenDroneMap/CloudODM).

## Support the Project

There are many ways to contribute back to the project:

 - ⭐️ us on GitHub.
 - Help us test the application.
 - Spread the word about OpenDroneMap on social media.
 - Help answer questions on the community [forum](https://community.opendronemap.org)
 - Become a contributor!




